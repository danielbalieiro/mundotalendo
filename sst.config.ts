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
        PK: "string",   // Partition key: EVENT#LEITURA#<uuid>, ERROR#<uuid>, APIKEY#*, WEBHOOK#PAYLOAD#<uuid>
        SK: "string",   // Sort key: COUNTRY#<iso3>, TIMESTAMP#*, KEY#*
        user: "string", // User name for GSI queries
      },
      primaryIndex: { hashKey: "PK", rangeKey: "SK" },
      globalIndexes: {
        UserIndex: {
          hashKey: "user",
          rangeKey: "PK", // Helps with efficient queries
          projection: "all",
        },
      },
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

    api.route("GET /users/locations", {
      handler: "packages/functions/users",
      runtime: "go",
      architecture: "arm64",
      link: [dataTable],
      timeout: "30 seconds",
      memory: "256 MB",
      transform: {
        function: (args) => {
          args.reservedConcurrentExecutions = 50; // Most called endpoint (same as stats)
        },
      },
    });

    // Next.js Frontend
    const web = new sst.aws.Nextjs("Web", {
      path: "./",
      environment: {
        NEXT_PUBLIC_API_URL: api.domain?.name
          ? `https://${api.domain.name}`
          : api.url,
        NEXT_PUBLIC_API_KEY: new sst.Secret("FrontendApiKey").value,
        NEXT_PUBLIC_SHOW_USER_MARKERS: "true", // User markers enabled for all stages
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

    // ============================================
    // MONITORING & ALERTING (PROD only)
    // ============================================

    const logRetentionDays = 14;
    const metricNamespace = "MundoTaLendo";
    const isProduction = $app.stage === "prod";

    // Log Groups with Retention
    const webhookLogGroup = new aws.cloudwatch.LogGroup("WebhookLogGroup", {
      name: `/aws/lambda/${$app.name}-${$app.stage}-webhook`,
      retentionInDays: logRetentionDays,
    });

    const statsLogGroup = new aws.cloudwatch.LogGroup("StatsLogGroup", {
      name: `/aws/lambda/${$app.name}-${$app.stage}-stats`,
      retentionInDays: logRetentionDays,
    });

    const usersLogGroup = new aws.cloudwatch.LogGroup("UsersLogGroup", {
      name: `/aws/lambda/${$app.name}-${$app.stage}-users`,
      retentionInDays: logRetentionDays,
    });

    const seedLogGroup = new aws.cloudwatch.LogGroup("SeedLogGroup", {
      name: `/aws/lambda/${$app.name}-${$app.stage}-seed`,
      retentionInDays: logRetentionDays,
    });

    const clearLogGroup = new aws.cloudwatch.LogGroup("ClearLogGroup", {
      name: `/aws/lambda/${$app.name}-${$app.stage}-clear`,
      retentionInDays: logRetentionDays,
    });

    // Metric Filters (PROD only)
    if (isProduction) {
    new aws.cloudwatch.LogMetricFilter("WebhookCountryNotFoundMetric", {
      logGroupName: webhookLogGroup.name,
      name: "CountryNotFoundErrors",
      pattern: '"Country not found:"',
      metricTransformation: {
        name: "CountryNotFoundErrorCount",
        namespace: metricNamespace,
        value: "1",
        defaultValue: "0",
        unit: "Count",
      },
    });

    new aws.cloudwatch.LogMetricFilter("WebhookUnmarshalErrorMetric", {
      logGroupName: webhookLogGroup.name,
      name: "UnmarshalErrors",
      pattern: '"Error parsing payload:"',
      metricTransformation: {
        name: "UnmarshalErrorCount",
        namespace: metricNamespace,
        value: "1",
        defaultValue: "0",
        unit: "Count",
      },
    });

    new aws.cloudwatch.LogMetricFilter("WebhookDynamoPutErrorMetric", {
      logGroupName: webhookLogGroup.name,
      name: "DynamoPutErrors",
      pattern: '"Error saving to DynamoDB:"',
      metricTransformation: {
        name: "DynamoPutErrorCount",
        namespace: metricNamespace,
        value: "1",
        defaultValue: "0",
        unit: "Count",
      },
    });

    new aws.cloudwatch.LogMetricFilter("WebhookPayloadTooLargeMetric", {
      logGroupName: webhookLogGroup.name,
      name: "PayloadTooLargeErrors",
      pattern: '"Payload too large:"',
      metricTransformation: {
        name: "PayloadTooLargeCount",
        namespace: metricNamespace,
        value: "1",
        defaultValue: "0",
        unit: "Count",
      },
    });

    new aws.cloudwatch.LogMetricFilter("StatsDynamoQueryErrorMetric", {
      logGroupName: statsLogGroup.name,
      name: "StatsQueryErrors",
      pattern: '"Error querying DynamoDB:"',
      metricTransformation: {
        name: "StatsQueryErrorCount",
        namespace: metricNamespace,
        value: "1",
        defaultValue: "0",
        unit: "Count",
      },
    });

    new aws.cloudwatch.LogMetricFilter("UsersDynamoQueryErrorMetric", {
      logGroupName: usersLogGroup.name,
      name: "UsersQueryErrors",
      pattern: '"Error querying DynamoDB:"',
      metricTransformation: {
        name: "UsersQueryErrorCount",
        namespace: metricNamespace,
        value: "1",
        defaultValue: "0",
        unit: "Count",
      },
    });

    new aws.cloudwatch.LogMetricFilter("WebhookAuthFailureMetric", {
      logGroupName: webhookLogGroup.name,
      name: "WebhookAuthFailures",
      pattern: '"Unauthorized: invalid API key"',
      metricTransformation: {
        name: "WebhookAuthFailureCount",
        namespace: metricNamespace,
        value: "1",
        defaultValue: "0",
        unit: "Count",
      },
    });

    new aws.cloudwatch.LogMetricFilter("StatsAuthFailureMetric", {
      logGroupName: statsLogGroup.name,
      name: "StatsAuthFailures",
      pattern: '"Unauthorized: invalid API key"',
      metricTransformation: {
        name: "StatsAuthFailureCount",
        namespace: metricNamespace,
        value: "1",
        defaultValue: "0",
        unit: "Count",
      },
    });

    new aws.cloudwatch.LogMetricFilter("UsersAuthFailureMetric", {
      logGroupName: usersLogGroup.name,
      name: "UsersAuthFailures",
      pattern: '"Unauthorized: invalid API key"',
      metricTransformation: {
        name: "UsersAuthFailureCount",
        namespace: metricNamespace,
        value: "1",
        defaultValue: "0",
        unit: "Count",
      },
    });

    // SNS Topic for Alarms
    const alarmTopic = new aws.sns.Topic("AlarmTopic", {
      displayName: `Mundo Tá Lendo - ${$app.stage.toUpperCase()} Alerts`,
    });

    new aws.sns.TopicSubscription("AlarmEmailSubscription", {
      topic: alarmTopic.arn,
      protocol: "email",
      endpoint: "daniel@balieiro.com",
    });

    // CloudWatch Alarms (Custom Metrics + DynamoDB)
    new aws.cloudwatch.MetricAlarm("DynamoPutErrorAlarm", {
      alarmName: `${$app.name}-${$app.stage}-webhook-dynamo-put-error`,
      alarmDescription: "CRITICAL: DynamoDB writes failing - DATA LOSS RISK",
      metricName: "DynamoPutErrorCount",
      namespace: metricNamespace,
      statistic: "Sum",
      period: 300,
      evaluationPeriods: 1,
      threshold: 1,
      comparisonOperator: "GreaterThanThreshold",
      treatMissingData: "notBreaching",
      actionsEnabled: true,
      alarmActions: [alarmTopic.arn],
    });

    new aws.cloudwatch.MetricAlarm("CountryNotFoundAlarm", {
      alarmName: `${$app.name}-${$app.stage}-webhook-country-not-found`,
      alarmDescription: "WARNING: Unmapped countries detected - add to mapping/countries.go",
      metricName: "CountryNotFoundErrorCount",
      namespace: metricNamespace,
      statistic: "Sum",
      period: 300,
      evaluationPeriods: 1,
      threshold: 1,
      comparisonOperator: "GreaterThanOrEqualToThreshold",
      treatMissingData: "notBreaching",
      actionsEnabled: true,
      alarmActions: [alarmTopic.arn],
    });

    new aws.cloudwatch.MetricAlarm("UnmarshalErrorAlarm", {
      alarmName: `${$app.name}-${$app.stage}-webhook-unmarshal-error`,
      alarmDescription: "WARNING: JSON parsing errors in webhook",
      metricName: "UnmarshalErrorCount",
      namespace: metricNamespace,
      statistic: "Sum",
      period: 300,
      evaluationPeriods: 1,
      threshold: 1,
      comparisonOperator: "GreaterThanOrEqualToThreshold",
      treatMissingData: "notBreaching",
      actionsEnabled: true,
      alarmActions: [alarmTopic.arn],
    });

    new aws.cloudwatch.MetricAlarm("StatsQueryErrorAlarm", {
      alarmName: `${$app.name}-${$app.stage}-stats-query-error`,
      alarmDescription: "CRITICAL: Stats endpoint DynamoDB query failures",
      metricName: "StatsQueryErrorCount",
      namespace: metricNamespace,
      statistic: "Sum",
      period: 300,
      evaluationPeriods: 1,
      threshold: 5,
      comparisonOperator: "GreaterThanThreshold",
      treatMissingData: "notBreaching",
      actionsEnabled: true,
      alarmActions: [alarmTopic.arn],
    });

    new aws.cloudwatch.MetricAlarm("UsersQueryErrorAlarm", {
      alarmName: `${$app.name}-${$app.stage}-users-query-error`,
      alarmDescription: "CRITICAL: Users endpoint DynamoDB query failures",
      metricName: "UsersQueryErrorCount",
      namespace: metricNamespace,
      statistic: "Sum",
      period: 300,
      evaluationPeriods: 1,
      threshold: 5,
      comparisonOperator: "GreaterThanThreshold",
      treatMissingData: "notBreaching",
      actionsEnabled: true,
      alarmActions: [alarmTopic.arn],
    });

    new aws.cloudwatch.MetricAlarm("DynamoReadThrottleAlarm", {
      alarmName: `${$app.name}-${$app.stage}-dynamo-read-throttle`,
      alarmDescription: "CRITICAL: DynamoDB reads being throttled - increase capacity",
      metricName: "UserErrors",
      namespace: "AWS/DynamoDB",
      dimensions: {
        TableName: dataTable.name,
      },
      statistic: "Sum",
      period: 300,
      evaluationPeriods: 1,
      threshold: 5,
      comparisonOperator: "GreaterThanThreshold",
      treatMissingData: "notBreaching",
      actionsEnabled: true,
      alarmActions: [alarmTopic.arn],
    });

    new aws.cloudwatch.MetricAlarm("DynamoWriteThrottleAlarm", {
      alarmName: `${$app.name}-${$app.stage}-dynamo-write-throttle`,
      alarmDescription: "CRITICAL: DynamoDB writes being throttled - increase capacity",
      metricName: "WriteThrottleEvents",
      namespace: "AWS/DynamoDB",
      dimensions: {
        TableName: dataTable.name,
      },
      statistic: "Sum",
      period: 300,
      evaluationPeriods: 1,
      threshold: 1,
      comparisonOperator: "GreaterThanThreshold",
      treatMissingData: "notBreaching",
      actionsEnabled: true,
      alarmActions: [alarmTopic.arn],
    });

    new aws.cloudwatch.MetricAlarm("AuthFailureAlarm", {
      alarmName: `${$app.name}-${$app.stage}-auth-failures`,
      alarmDescription: "SECURITY: High auth failures - possible brute force attack",
      metricName: "WebhookAuthFailureCount",
      namespace: metricNamespace,
      statistic: "Sum",
      period: 300,
      evaluationPeriods: 1,
      threshold: 20,
      comparisonOperator: "GreaterThanThreshold",
      treatMissingData: "notBreaching",
      actionsEnabled: true,
      alarmActions: [alarmTopic.arn],
    });

    // CloudWatch Dashboard
    const dashboard = new aws.cloudwatch.Dashboard("ProductionDashboard", {
      dashboardName: `${$app.name}-${$app.stage}-dashboard`,
      dashboardBody: $jsonStringify({
        widgets: [
          {
            type: "metric",
            x: 0,
            y: 0,
            width: 12,
            height: 6,
            properties: {
              title: "API Error Rates (5 min)",
              metrics: [
                ["AWS/Lambda", "Errors", { stat: "Sum", label: "Total Errors", color: "#d62728" }],
              ],
              period: 300,
              region: "us-east-2",
              yAxis: { left: { min: 0 } },
            },
          },
          {
            type: "metric",
            x: 12,
            y: 0,
            width: 12,
            height: 6,
            properties: {
              title: "API Latency (Average)",
              metrics: [
                ["AWS/Lambda", "Duration", { stat: "Average", label: "Avg Duration", color: "#1f77b4" }],
              ],
              period: 300,
              region: "us-east-2",
              yAxis: { left: { label: "Milliseconds", min: 0 } },
            },
          },
          {
            type: "metric",
            x: 0,
            y: 6,
            width: 12,
            height: 6,
            properties: {
              title: "Webhook Error Breakdown",
              metrics: [
                [metricNamespace, "CountryNotFoundErrorCount", { stat: "Sum", color: "#ff7f0e" }],
                ["...", "UnmarshalErrorCount", { stat: "Sum", color: "#d62728" }],
                ["...", "DynamoPutErrorCount", { stat: "Sum", color: "#9467bd" }],
              ],
              period: 300,
              region: "us-east-2",
              stacked: true,
            },
          },
          {
            type: "metric",
            x: 12,
            y: 6,
            width: 12,
            height: 6,
            properties: {
              title: "DynamoDB Health",
              metrics: [
                ["AWS/DynamoDB", "UserErrors", "TableName", dataTable.name],
              ],
              period: 300,
              region: "us-east-2",
              stat: "Sum",
              yAxis: { left: { min: 0 } },
            },
          },
        ],
      }),
    });
    } // End of isProduction monitoring block

    return {
      api: api.url,
      web: web.url,
      dataTable: dataTable.name,
    };
  },
});
