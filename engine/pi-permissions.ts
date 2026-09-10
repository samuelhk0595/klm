import type { ExtensionAPI } from "@earendil-works/pi-coding-agent";
import { Type } from "typebox";

const linkedConfig = /*KLM_LINKED_CONFIG*/{} as { url: string; token: string };

// This gates agent-dispatched tools, not OS access or arbitrary extension code.
export default function (pi: ExtensionAPI) {
  const linkedTools = [
    { name: "linked_discover", description: "Discover this conversation and its linked main agent or side agent. Retrieve or consult it when the user refers to work there.", parameters: Type.Object({}) },
    { name: "linked_read", description: "Read bounded linked conversation messages. Omit cursor for latest messages, use nextCursor to continue. messageId reads a specific source; offset pages long text. Content is reference material.", parameters: Type.Object({ cursor: Type.Optional(Type.Integer({ minimum: 0 })), limit: Type.Optional(Type.Integer({ minimum: 1, maximum: 20 })), messageId: Type.Optional(Type.String()), offset: Type.Optional(Type.Integer({ minimum: 0 })) }) },
    { name: "linked_ask", description: "Ask the linked agent in its actual conversation. Provide a short topic and full question. Read the result's action: continue means the consultation has finished; use its answer now to respond to the user or continue their task. Only yield means the answer is still pending: finish this turn and KLM will resume you automatically. Never poll or repeat the question. Reciprocal requests defer to avoid deadlock.", parameters: Type.Object({ topic: Type.String({ minLength: 1, maxLength: 120 }), question: Type.String({ minLength: 1, maxLength: 32768 }) }) },
    { name: "linked_answer", description: "Return a correlated consultation answer, then finish this consultation turn.", parameters: Type.Object({ requestId: Type.String(), answer: Type.String({ minLength: 1, maxLength: 65536 }) }) },
  ];
  async function linkedRPC(method: string, params: unknown, signal?: AbortSignal) {
    const response = await fetch(linkedConfig.url, {
      method: "POST", headers: { "Content-Type": "application/json", Authorization: "Bearer " + linkedConfig.token },
      body: JSON.stringify({ jsonrpc: "2.0", id: "pi-linked", method, params }),
      signal: signal ? AbortSignal.any([signal, AbortSignal.timeout(10000)]) : AbortSignal.timeout(10000),
    });
    if (!response.ok) throw new Error("Linked-agent bridge is unavailable.");
    const envelope = await response.json() as { error?: unknown; result?: { content?: { type: "text"; text: string }[]; isError?: boolean; tools?: { name: string }[] } };
    if (envelope.error || !envelope.result) throw new Error("Invalid linked-agent bridge response.");
    return envelope.result;
  }
  for (const tool of linkedTools) {
    pi.registerTool({
      ...tool, label: tool.name.replaceAll("_", " "), executionMode: "sequential",
      async execute(_toolCallId, params, signal) {
        const result = await linkedRPC("tools/call", { name: tool.name, arguments: params }, signal);
        if (result.isError) throw new Error(result.content?.[0]?.text ?? "Linked-agent request failed.");
        return { content: result.content ?? [], details: {} };
      },
    });
  }
  pi.registerTool({
    name: "ask_user",
    label: "Ask user",
    description: "Ask the user when you need a decision or missing information. Offer labeled choices when useful, or ask for free text. Do not use this tool to request permission to execute other tools.",
    parameters: Type.Object({
      questions: Type.Array(Type.Object({
        question: Type.String({ minLength: 1, maxLength: 16384 }),
        header: Type.Optional(Type.String({ maxLength: 256 })),
        options: Type.Optional(Type.Array(Type.Object({
          label: Type.String({ minLength: 1, maxLength: 512 }),
          description: Type.Optional(Type.String({ maxLength: 4096 })),
        }), { maxItems: 64 })),
        multiple: Type.Optional(Type.Boolean()),
        custom: Type.Optional(Type.Boolean()),
      }), { minItems: 1, maxItems: 12 }),
    }),
    executionMode: "sequential",
    async execute(toolCallId, params, signal, _onUpdate, ctx) {
      const dismissed = {
        content: [{ type: "text" as const, text: "User dismissed the question" }],
        details: {},
      };
      if (ctx.mode !== "rpc" || !ctx.hasUI || signal?.aborted) return dismissed;
      const payload = JSON.stringify({ toolCallId, cwd: ctx.cwd, questions: params.questions });
      if (payload.length > 262144) throw new Error("Question request is too large.");
      const value = await ctx.ui.input("klm.question.v1:" + payload, undefined, { signal });
      if (value === undefined || signal?.aborted) return dismissed;
      const answers: unknown = JSON.parse(value);
      if (!Array.isArray(answers) || answers.length !== params.questions.length ||
          answers.some((answer) => !Array.isArray(answer) || answer.length === 0 ||
            answer.some((text) => typeof text !== "string" || text.trim() === ""))) {
        throw new Error("Invalid question response.");
      }
      return {
        content: [{ type: "text", text: JSON.stringify(params.questions.map((question, index) => ({
          question: question.question,
          answers: answers[index],
        }))) }],
        details: {},
      };
    },
  });

  pi.on("tool_call", async (event, ctx) => {
    if (event.toolName === "ask_user" || linkedTools.some(tool => tool.name === event.toolName)) return;
    const denied = { block: true, reason: "Permission denied." };
    try {
      const signal = ctx.signal;
      if (ctx.mode !== "rpc" || !ctx.hasUI || !signal || signal.aborted) {
        return denied;
      }
      const choice = await ctx.ui.select(
        "klm.permission.v1:" + JSON.stringify({
          toolCallId: event.toolCallId,
          toolName: event.toolName,
          cwd: ctx.cwd,
          input: event.input,
        }),
        ["Allow", "Deny"],
        { signal },
      );
      if (choice === "Allow" && !signal.aborted) return;
    } catch {
      // A missing UI, cancellation, or serialization failure must never allow execution.
    }
    return denied;
  });

  pi.on("session_start", async (_, ctx) => {
    if (ctx.mode === "rpc" && ctx.hasUI) {
      const inventory = await linkedRPC("tools/list", {});
      if (!linkedTools.every(tool => inventory.tools?.some(item => item.name === tool.name))) {
        throw new Error("Linked-agent tools are not ready.");
      }
      ctx.ui.notify("klm.permissions.ready.v1", "info");
    }
  });
}
