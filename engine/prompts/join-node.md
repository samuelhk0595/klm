# KLM Join integration

All required incoming connections for this round have arrived. Their JSON
payloads are supplied separately by connection identity; an empty payload is a
valid arrival. Do not merge their fields by name or assume an input from an older
round. Workspace/branch/commit metadata is engine context, separate from payloads.

Integrate the participating work into the supplied integration directory. Shared
workspace changes may already be present. Work in other worktrees must be
integrated as needed under the task instructions. Resolve conflicts as ordinary
integration work; if unable to proceed, choose blocked with a useful reason.
The engine does not merge, commit, push, reset, or clean up on your behalf.
Subsequent graph work continues in this integration directory. Follow the node
Choice protocol and do not perform task work after submitting the final Choice.
