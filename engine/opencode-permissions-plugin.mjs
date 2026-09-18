// Owned tool preflight. Native restrictions still apply outside YOLO mode.
const permissionConfig = /*KLM_PERMISSION_CONFIG*/{};
export default async function (context) {
  const hooks = await graphPlugin(context);
  return {
    ...hooks,
    "tool.execute.before": async (input, output) => {
      const response = await fetch(permissionConfig.url, {
        method: "POST",
        headers: { "Content-Type": "application/json", Authorization: "Bearer " + permissionConfig.token },
        body: JSON.stringify({ jsonrpc: "2.0", id: crypto.randomUUID(), method: "klm/permission/opencode", params: {
          sessionID: input.sessionID, toolCallId: input.callID, toolName: input.tool,
          cwd: context.directory, input: output.args,
        } }),
      });
      if (!response.ok) throw new Error("KLM permission gate is unavailable.");
      const envelope = await response.json();
      if (envelope.error || envelope.result?.allowed !== true) throw new Error("Permission denied by KLM.");
      // Admission happens after permission resolution; a sealed graph still denies.
      await hooks["tool.execute.before"]?.(input, output);
    },
    config: async (config) => {
      if (!permissionConfig.yolo) return;
      config.permission = { "*": "allow" };
      for (const agent of Object.values(config.agent ?? {})) {
        agent.permission = { "*": "allow" };
        if (agent.tools) {
          for (const tool of Object.keys(agent.tools)) agent.tools[tool] = true;
        }
      }
      if (config.tools) {
        for (const tool of Object.keys(config.tools)) config.tools[tool] = true;
      }
    },
  };
}
