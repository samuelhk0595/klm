package main

import "maps"

func runnerManifestTools(manifest []map[string]any) []map[string]any {
	tools := append([]map[string]any{}, manifest...)
	for i, tool := range tools {
		schema := object(tool["inputSchema"])
		if required, exists := schema["required"]; exists && required == nil {
			// Legacy captures encoded an empty required list as null. Repair only
			// the harness projection; immutable captures and receipts stay intact.
			projected := maps.Clone(tool)
			projectedSchema := maps.Clone(schema)
			projectedSchema["required"] = []string{}
			projected["inputSchema"] = projectedSchema
			tools[i] = projected
		}
	}
	return tools
}
