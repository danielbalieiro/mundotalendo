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

    // S3 Bucket for webhook payloads (cheaper than DynamoDB for large payloads)
    const payloadBucket = new sst.aws.Bucket("PayloadBucket", {
      transform: {
        bucket: (args) => {
          args.lifecycleRules = [
            {
              id: "expire-old-payloads",
              enabled: true,
              expirations: [{ days: 90 }],
            },
          ];
        },
      },
    });

    // Dead Letter Queue for failed webhook messages
    const webhookDLQ = new sst.aws.Queue("WebhookDLQ", {
      transform: {
        queue: (args) => {
          args.messageRetentionSeconds = 1209600; // 14 days
        },
      },
    });

    // Main webhook processing queue
    const webhookQueue = new sst.aws.Queue("WebhookQueue", {
      dlq: {
        queue: webhookDLQ.arn,
        retry: 3, // 3 retries before DLQ
      },
      transform: {
        queue: (args) => {
          args.visibilityTimeoutSeconds = 90; // Must be >= consumer timeout
          args.messageRetentionSeconds = 345600; // 4 days
        },
      },
    });

    // Consumer Lambda - processes messages from SQS
    webhookQueue.subscribe({
      handler: "packages/functions/consumer",
      runtime: "go",
      architecture: "arm64",
      link: [dataTable, payloadBucket],
      timeout: "60 seconds",
      memory: "512 MB",
      transform: {
        function: (args) => {
          args.reservedConcurrentExecutions = 10;
        },
      },
    }, {
      batch: { size: 1 },
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
      link: [dataTable, payloadBucket, webhookQueue],
      timeout: "10 seconds", // Reduced - only validation and queueing
      memory: "128 MB",      // Reduced - less processing
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

    api.route("POST /migrate", {
      handler: "packages/functions/migrate",
      runtime: "go",
      architecture: "arm64",
      link: [dataTable],
      timeout: "300 seconds", // 5 minutes for large migrations
      memory: "1024 MB", // More memory for scanning large tables
      transform: {
        function: (args) => {
          args.reservedConcurrentExecutions = 1; // Only one migration at a time
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

    api.route("GET /readings/{iso3}", {
      handler: "packages/functions/readings",
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

    const consumerLogGroup = new aws.cloudwatch.LogGroup("ConsumerLogGroup", {
      name: `/aws/lambda/${$app.name}-${$app.stage}-consumer`,
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

    // DLQ Alarm - Critical: Messages failing all retries
    new aws.cloudwatch.MetricAlarm("WebhookDLQAlarm", {
      alarmName: `${$app.name}-${$app.stage}-webhook-dlq-messages`,
      alarmDescription: "CRITICAL: Webhook messages failing all retries - investigate immediately",
      metricName: "ApproximateNumberOfMessagesVisible",
      namespace: "AWS/SQS",
      dimensions: {
        QueueName: webhookDLQ.name,
      },
      statistic: "Sum",
      period: 60,
      evaluationPeriods: 1,
      threshold: 1,
      comparisonOperator: "GreaterThanOrEqualToThreshold",
      treatMissingData: "notBreaching",
      actionsEnabled: true,
      alarmActions: [alarmTopic.arn],
    });

    // Consumer error metrics
    new aws.cloudwatch.LogMetricFilter("ConsumerProcessingErrorMetric", {
      logGroupName: consumerLogGroup.name,
      name: "ConsumerProcessingErrors",
      pattern: '"ERROR"',
      metricTransformation: {
        name: "ConsumerProcessingErrorCount",
        namespace: metricNamespace,
        value: "1",
        defaultValue: "0",
        unit: "Count",
      },
    });

    new aws.cloudwatch.LogMetricFilter("ConsumerS3FetchErrorMetric", {
      logGroupName: consumerLogGroup.name,
      name: "ConsumerS3FetchErrors",
      pattern: '"Error fetching payload from S3"',
      metricTransformation: {
        name: "ConsumerS3FetchErrorCount",
        namespace: metricNamespace,
        value: "1",
        defaultValue: "0",
        unit: "Count",
      },
    });

    // Lambda Panic/Crash metric filters - catches nil pointer, panics, runtime errors
    new aws.cloudwatch.LogMetricFilter("WebhookPanicMetric", {
      logGroupName: webhookLogGroup.name,
      name: "WebhookPanics",
      pattern: '?"panic" ?"nil pointer dereference" ?"runtime error"',
      metricTransformation: {
        name: "WebhookPanicCount",
        namespace: metricNamespace,
        value: "1",
        defaultValue: "0",
        unit: "Count",
      },
    });

    new aws.cloudwatch.LogMetricFilter("ConsumerPanicMetric", {
      logGroupName: consumerLogGroup.name,
      name: "ConsumerPanics",
      pattern: '?"panic" ?"nil pointer dereference" ?"runtime error"',
      metricTransformation: {
        name: "ConsumerPanicCount",
        namespace: metricNamespace,
        value: "1",
        defaultValue: "0",
        unit: "Count",
      },
    });

    // Lambda Panic Alarms - CRITICAL: Catches crashes before any logging
    new aws.cloudwatch.MetricAlarm("WebhookPanicAlarm", {
      alarmName: `${$app.name}-${$app.stage}-webhook-panic`,
      alarmDescription: "CRITICAL: Webhook Lambda is crashing/panicking - immediate attention required",
      metricName: "WebhookPanicCount",
      namespace: metricNamespace,
      statistic: "Sum",
      period: 60,
      evaluationPeriods: 1,
      threshold: 1,
      comparisonOperator: "GreaterThanOrEqualToThreshold",
      treatMissingData: "notBreaching",
      actionsEnabled: true,
      alarmActions: [alarmTopic.arn],
    });

    new aws.cloudwatch.MetricAlarm("ConsumerPanicAlarm", {
      alarmName: `${$app.name}-${$app.stage}-consumer-panic`,
      alarmDescription: "CRITICAL: Consumer Lambda is crashing/panicking - webhooks NOT being processed",
      metricName: "ConsumerPanicCount",
      namespace: metricNamespace,
      statistic: "Sum",
      period: 60,
      evaluationPeriods: 1,
      threshold: 1,
      comparisonOperator: "GreaterThanOrEqualToThreshold",
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
      payloadBucket: payloadBucket.name,
      webhookQueue: webhookQueue.url,
      webhookDLQ: webhookDLQ.url,
    };
  },
});
