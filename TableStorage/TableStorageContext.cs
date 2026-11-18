namespace AzurePatterns.TableStorage
{
    using System;
    using System.Collections.Generic;
    using System.Linq;
    using System.Net;

    using Azure;
    using Azure.Data.Tables;

    using AzurePatterns.Interfaces;

    // Unit of Work over Azure Tables: changes are queued and sent as
    // entity group transactions (one per table and partition) on Commit.
    public class TableStorageContext : IUnitOfWork
    {
        private const int MaxTransactionSize = 100;

        private readonly List<(string Table, TableTransactionAction Action)> pending =
            new List<(string, TableTransactionAction)>();

        // Constructor allows for setting up a specific connection string (for testing).
        // Otherwise it is read from the CloudConnectionString environment variable.
        public TableStorageContext(string connectionString = null)
        {
            var cloudConnectionString = connectionString
                                        ?? Environment.GetEnvironmentVariable("CloudConnectionString");

            if (string.IsNullOrEmpty(cloudConnectionString))
            {
                throw new InvalidOperationException(
                    "Expected a connection string or the CloudConnectionString environment variable.");
            }

            this.Service = new TableServiceClient(cloudConnectionString);
        }

        public TableServiceClient Service { get; private set; }

        public void Enqueue(string table, TableTransactionAction action)
        {
            this.pending.Add((table, action));
        }

        // NOTE: the implementation of Commit may vary depending on your desired table behaviour.
        public void Commit()
        {
            var batches = this.pending
                .GroupBy(p => (p.Table, p.Action.Entity.PartitionKey))
                .SelectMany(g => g.Select((p, i) => (p, i)).GroupBy(x => x.i / MaxTransactionSize))
                .Select(g => g.Select(x => x.p).ToList())
                .ToList();

            foreach (var batch in batches)
            {
                var client = this.Service.GetTableClient(batch[0].Table);

                try
                {
                    client.SubmitTransaction(batch.Select(b => b.Action));
                }
                catch (RequestFailedException exception) when (exception.Status == (int)HttpStatusCode.Conflict)
                {
                    // a conflict may arise on a retry where it succeeded so this is ignored.
                }
            }

            this.pending.Clear();
        }

        public void Rollback()
        {
            this.pending.Clear();
        }
    }
}
