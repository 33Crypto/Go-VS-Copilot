namespace AzurePatterns.TableStorage
{
    using System;

    using Azure.Data.Tables;

    using AzurePatterns.Interfaces;

    public class RepositoryBase
    {
        public RepositoryBase(IUnitOfWork context, string table)
        {
            if (context == null)
            {
                throw new ArgumentNullException("context");
            }

            if (string.IsNullOrEmpty(table))
            {
                throw new ArgumentNullException("table", "Expected a table name.");
            }

            this.Context = context as TableStorageContext;
            this.Table = table;

            // belt-and-braces code - ensure the table is there for the repository.
            if (this.Context != null)
            {
                this.Client = this.Context.Service.GetTableClient(this.Table);
                this.Client.CreateIfNotExists();
            }
        }

        protected TableStorageContext Context { get; private set; }

        protected TableClient Client { get; private set; }

        protected string Table { get; private set; }
    }
}
