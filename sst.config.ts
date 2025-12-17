/// <reference path="./.sst/platform/config.d.ts" />

export default $config({
  app(input) {
    return {
      name: "mundotalendo",
      removal: input?.stage === "production" ? "retain" : "remove",
      home: "aws",
      providers: {
        aws: {
          region: "us-east-2",
        },
      },
    };
  },
  async run() {
    // DynamoDB Table for reading telemetry
    const table = new sst.aws.Dynamo("Leituras", {
      fields: {
        PK: "string",
        SK: "string",
      },
      primaryIndex: { hashKey: "PK", rangeKey: "SK" },
    });

    // DynamoDB Table for failed webhooks
    const falhasTable = new sst.aws.Dynamo("Falhas", {
      fields: {
        PK: "string",  // "ERROR#<errorType>"
        SK: "string",  // "TIMESTAMP#<RFC3339>"
      },
      primaryIndex: { hashKey: "PK", rangeKey: "SK" },
    });

    // API Gateway with inline route definitions
    const api = new sst.aws.ApiGatewayV2("Api", {
      cors: {
        allowOrigins: ["*"],
        allowMethods: ["GET", "POST", "OPTIONS"],
        allowHeaders: ["Content-Type", "Authorization"],
      },
      domain:
        $app.stage === "prod"
          ? {
              name: "api.mundotalendo.com.br",
              dns: sst.aws.dns(),
            }
          : $app.stage === "dev"
            ? {
                name: "api.dev.mundotalendo.com.br",
                dns: sst.aws.dns(),
              }
            : undefined,
    });

    api.route("POST /webhook", {
      handler: "packages/functions/webhook",
      runtime: "go",
      architecture: "arm64",
      link: [table, falhasTable],
      timeout: "30 seconds",
      memory: "256 MB",
    });

    api.route("POST /test/seed", {
      handler: "packages/functions/seed",
      runtime: "go",
      architecture: "arm64",
      link: [table],
      timeout: "30 seconds",
      memory: "256 MB",
    });

    api.route("GET /stats", {
      handler: "packages/functions/stats",
      runtime: "go",
      architecture: "arm64",
      link: [table],
      timeout: "30 seconds",
      memory: "256 MB",
    });

    api.route("POST /clear", {
      handler: "packages/functions/clear",
      runtime: "go",
      architecture: "arm64",
      link: [table, falhasTable],
      timeout: "60 seconds",
      memory: "512 MB",
    });

    // Next.js Frontend
    const web = new sst.aws.Nextjs("Web", {
      path: "./",
      environment: {
        NEXT_PUBLIC_API_URL: api.url,
      },
      domain:
        $app.stage === "prod"
          ? {
              name: "mundotalendo.com.br",
              dns: sst.aws.dns(),
            }
          : $app.stage === "dev"
            ? {
                name: "dev.mundotalendo.com.br",
                dns: sst.aws.dns(),
              }
            : undefined,
    });

    return {
      api: api.url,
      web: web.url,
      table: table.name,
    };
  },
});
