import { useEffect, useMemo, useRef, useState } from 'react';
import { Background, MarkerType, Panel, ReactFlow } from '@xyflow/react';
import type { Agent } from '../agents/demo';
import type { Graph } from './demo';
import { resolveAgentNode } from './agent-node';
import { CanvasControls, graphViewNodeTypes, type CanvasNode } from './GraphCanvas';
import { graphDrawings } from './graph-view-demo';
import { GraphRunActivity } from './GraphRunActivity';
import { getNodeActivity } from './graph-run-demo';
import { NodeConfigurationPanel } from './NodeConfigurationPanel';

export function GraphView({ graph, agents, visible, onRunningChange }: { graph: Graph; agents: Agent[]; visible: boolean; onRunningChange: (running: boolean) => void }) {
  const root = useRef<HTMLElement>(null);
  const [inspectingId, setInspectingId] = useState<string | null>(null);
  const [step, setStep] = useState(0);
  const [hasOpened, setHasOpened] = useState(visible);
  useEffect(() => { if (visible) setHasOpened(true); }, [visible]);
  const drawing = graphDrawings[graph.id];
  useEffect(() => {
    if (!hasOpened || graph.id !== 'csv-export') return;
    // Finish planning once, pass through ready, then keep implementation active.
    const timer = window.setInterval(() => setStep(current => current < 15 ? current + 1 : 9), 3000);
    return () => window.clearInterval(timer);
  }, [graph.id, hasOpened]);
  function closeInspection() {
    root.current?.querySelector<HTMLButtonElement>(`[data-id="${CSS.escape(inspectingId ?? '')}"] .graph-node-inspect-button`)?.focus();
    setInspectingId(null);
  }
  const running = hasOpened && !!drawing?.nodes.some(node => getNodeActivity(graph.id, node.id, step).status === 'running');
  const nodes = useMemo<CanvasNode[]>(() => (drawing?.nodes ?? []).map(node => {
    const agent = agents.find(item => item.id === (node.data.join?.agentId ?? node.data.agentId));
    const status = getNodeActivity(graph.id, node.id, step).status;
    return { ...node, draggable: false, selectable: false, connectable: false, deletable: false, focusable: false,
      data: { ...node.data, readOnly: true, agent, settings: agent ? resolveAgentNode(agent, node.data.overrides ?? {}) : undefined,
        running: running && status === 'running', activityStatus: running ? status : undefined,
        inspectionMode: running ? 'activity' : 'configuration', configuring: !running && node.id === inspectingId,
        inspecting: node.id === inspectingId, onInspect: () => setInspectingId(node.id) },
    };
  }), [drawing, agents, graph.id, step, inspectingId, running]);
  useEffect(() => {
    onRunningChange(running);
    return () => onRunningChange(false);
  }, [running, onRunningChange]);
  const inspectedNode = nodes.find(node => node.id === inspectingId);
  const inspectedName = inspectedNode?.data.name ?? inspectedNode?.data.choice?.name ?? inspectedNode?.data.fork?.name ?? inspectedNode?.data.terminal?.name ?? (inspectedNode?.type === 'join' ? `Join · ${inspectedNode.data.agent?.name ?? 'Integration'}` : 'Node');

  return <section ref={root} hidden={!visible} className="graph-canvas graph-canvas--readonly" aria-label={`${graph.name} graph canvas`} onKeyDown={event => {
    if (event.key === 'Escape' && inspectingId) { event.stopPropagation(); closeInspection(); }
  }}>
    {hasOpened && <ReactFlow<CanvasNode> nodes={nodes} edges={drawing?.edges ?? []} nodeTypes={graphViewNodeTypes}
      nodesDraggable={false} nodesConnectable={false} nodesFocusable={false}
      edgesReconnectable={false} edgesFocusable={false} elementsSelectable={false}
      deleteKeyCode={null} selectionKeyCode={null} multiSelectionKeyCode={null}
      panOnDrag zoomOnScroll zoomOnPinch minZoom={0.1} maxZoom={2}
      fitView fitViewOptions={{ maxZoom: 1, padding: 0.15, nodes: graph.id === 'csv-export' ? [{ id: 'plan' }, { id: 'ready' }, { id: 'implement' }] : undefined }}
      defaultEdgeOptions={{ type: 'smoothstep', markerEnd: { type: MarkerType.ArrowClosed, color: 'var(--color-muted)' }, style: { stroke: 'var(--color-muted)' } }}>
      <Background gap={24} size={1} color="var(--color-faint)" />
      <CanvasControls />
      {inspectedNode && (running
        ? <GraphRunActivity key={inspectedNode.id} nodeName={inspectedName} activity={getNodeActivity(graph.id, inspectedNode.id, step)} onClose={closeInspection} />
        : <Panel position="center-right" className="graph-agent-panel-position"><NodeConfigurationPanel key={inspectedNode.id} node={inspectedNode} nodes={nodes} edges={drawing?.edges ?? []} onClose={closeInspection} /></Panel>)}
    </ReactFlow>}
  </section>;
}
