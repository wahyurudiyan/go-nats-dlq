# NATS Jetstream Dead Queue Letter

## 📘 Diagram Explanation: JetStream DLQ Flow
- The Service publishes a message to a subject (e.g., jobs.taskA) in the JetStream stream called jobs.

- NATS JetStream receives the message and delivers it to the Worker via a configured consumer (e.g., job_consumer).

- The Worker is expected to process the message and send an ACK back to NATS if successful.

- If the Worker fails to process the message (e.g., crashes, logic error, or does not ACK), JetStream will redeliver the message up to a configured max_deliver count (e.g., 5 times).

- After exceeding the maximum delivery attempts without a successful ACK, JetStream emits an advisory event to the subject:

```bash
$JS.EVENT.ADVISORY.CONSUMER.MAX_DELIVERIES.jobs.job_consumer
```

- This event includes metadata about the failed message, such as subject, sequence number, delivery count, and more.

- The advisory is captured by a separate stream — the DLQ Stream — which stores these events for further inspection, retry logic, or alerting.

## 🔍 Purpose of the DLQ Stream
- Acts as a holding area for messages that consistently fail processing.

- Helps operators and developers monitor failure patterns.

- Allows for manual or automated replay of failed messages.

- Integrates with logging, alerting, or incident management systems.

## Jetstream Setup

At the first, we need to define stream named `jobs`.

```bash
❯ nats stream -s localhost:4222 create jobs                                                                                                         
? Subjects jobs.*.*
? Storage file
? Replication 1
? Retention Policy Work Queue
? Discard Policy New
? Stream Messages Limit 50000
? Per Subject Messages Limit -1
? Total Stream Size 256MB
? Message TTL 7d
? Max Message Size -1
? Duplicate tracking time window 2m0s
? Allow message Roll-ups No
? Allow message deletion Yes
? Allow purging subjects or the entire stream Yes
Stream jobs was created

Information for Stream jobs created 2025-06-12 08:59:57

              Subjects: jobs.*.*
              Replicas: 1
               Storage: File

Options:

             Retention: WorkQueue
       Acknowledgments: true
        Discard Policy: New
      Duplicate Window: 2m0s
            Direct Get: true
     Allows Msg Delete: true
          Allows Purge: true
        Allows Rollups: false

Limits:

      Maximum Messages: 50,000
   Maximum Per Subject: unlimited
         Maximum Bytes: 256 MiB
           Maximum Age: 7d0h0m0s
  Maximum Message Size: unlimited
     Maximum Consumers: unlimited

Metadata:

           _nats.level: 1
       _nats.req.level: 0
             _nats.ver: 2.11.4

State:

              Messages: 0
                 Bytes: 0 B
        First Sequence: 0
         Last Sequence: 0
      Active Consumers: 0
```

Next, we need to add another job that will act as Dead Letter Queue (DLQ).

```bash
nats stream -s localhost:4222 create jobs_dlq                                                                                                        ─╯
? Subjects $JS.EVENT.ADVISORY.CONSUMER.MAX_DELIVERIES.jobs.*
? Storage file
? Replication 1
? Retention Policy Limits
? Discard Policy Old
? Stream Messages Limit 50000
? Per Subject Messages Limit -1
? Total Stream Size 1GB
? Message TTL 7d
? Max Message Size -1
? Duplicate tracking time window 2m0s
? Allow message Roll-ups No
? Allow message deletion Yes
? Allow purging subjects or the entire stream Yes
Stream jobs_dlq was created

Information for Stream jobs_dlq created 2025-06-12 09:49:53

              Subjects: $JS.EVENT.ADVISORY.CONSUMER.MAX_DELIVERIES.jobs.*
              Replicas: 1
               Storage: File

Options:

             Retention: Limits
       Acknowledgments: true
        Discard Policy: Old
      Duplicate Window: 2m0s
            Direct Get: true
     Allows Msg Delete: true
          Allows Purge: true
        Allows Rollups: false

Limits:

      Maximum Messages: 50,000
   Maximum Per Subject: unlimited
         Maximum Bytes: 1.0 GiB
           Maximum Age: 7d0h0m0s
  Maximum Message Size: unlimited
     Maximum Consumers: unlimited

Metadata:

           _nats.level: 1
       _nats.req.level: 0
             _nats.ver: 2.11.4

State:

              Messages: 0
                 Bytes: 0 B
        First Sequence: 0
         Last Sequence: 0
      Active Consumers: 0
```