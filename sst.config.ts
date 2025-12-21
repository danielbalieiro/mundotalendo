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
    // DynamoDB Single Table for all data (events, errors, API keys)
    const dataTable = new sst.aws.Dynamo("DataTable", {
      fields: {
        PK: "string",  // Partition key: EVENT#*, ERROR#*, APIKEY#*
        SK: "string",  // Sort key: TIMESTAMP#*, KEY#*
      },
      primaryIndex: { hashKey: "PK", rangeKey: "SK" },
      transform: {
        table: {
          pointInTimeRecovery: { enabled: true },
        },
      },
    });

    // API Gateway with inline route definitions
    const api = new sst.aws.ApiGatewayV2("Api", {
      cors: {
        allowOrigins: [
          "https://mundotalendo.com.br",
          "https://dev.mundotalendo.com.br",
          "http://localhost:3000", // Local development
        ],
        allowMethods: ["GET", "POST", "OPTIONS"],
        allowHeaders: ["Content-Type", "Authorization", "X-API-Key"],
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
      link: [dataTable],
      timeout: "30 seconds",
      memory: "256 MB",
      transform: {
        function: (args) => {
          args.reservedConcurrentExecutions = 10;
        },
      },
    });

    api.route("POST /test/seed", {
      handler: "packages/functions/seed",
      runtime: "go",
      architecture: "arm64",
      link: [dataTable],
      timeout: "30 seconds",
      memory: "256 MB",
      transform: {
        function: (args) => {
          args.reservedConcurrentExecutions = 5;
        },
      },
    });

    api.route("GET /stats", {
      handler: "packages/functions/stats",
      runtime: "go",
      architecture: "arm64",
      link: [dataTable],
      timeout: "30 seconds",
      memory: "256 MB",
      transform: {
        function: (args) => {
          args.reservedConcurrentExecutions = 50; // Most called endpoint
        },
      },
    });

    api.route("POST /clear", {
      handler: "packages/functions/clear",
      runtime: "go",
      architecture: "arm64",
      link: [dataTable],
      timeout: "60 seconds",
      memory: "512 MB",
      transform: {
        function: (args) => {
          args.reservedConcurrentExecutions = 5;
        },
      },
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
      dataTable: dataTable.name,
    };
  },
});
