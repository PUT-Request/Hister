/// <reference path="../globals.d.ts" />

export default defineTool({
  id: "example-tool",
  description: "A starter Open Dynamic Workflow tool that echoes input.",
  inputSchema: {
    type: "object",
    properties: {
      message: { type: "string" }
    },
    required: ["message"]
  },
  outputSchema: {
    type: "object",
    properties: {
      echo: { type: "string" }
    },
    required: ["echo"]
  },
  run(input: { message: string }, context) {
    context.log("Executing example-tool", { message: input.message });
    return {
      echo: input.message
    };
  }
});
