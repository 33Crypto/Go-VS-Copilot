namespace AzurePatterns.Repositories
{
    using System.Collections.Generic;
    using System.Linq;

    using Azure.Data.Tables;

    using AzurePatterns.Entities;
    using AzurePatterns.Interfaces;
    using AzurePatterns.TableStorage;

    // NOTE: It's an EntityRepository not a TableRepository because there
    // may be more than one kind of entity stored in a single table.
    public class EntityRepository : RepositoryBase, IEntityRepository
    {
        public EntityRepository(IUnitOfWork context) : base(context, "table")
        {
        }

        public void Save(Entity entity)
        {
            // Insert or Merge Entity aka Upsert, sent on Commit.
            this.Context.Enqueue(this.Table, new TableTransactionAction(TableTransactionActionType.UpsertMerge, entity));
        }

        public void Delete(Entity entity)
        {
            this.Context.Enqueue(this.Table, new TableTransactionAction(TableTransactionActionType.Delete, entity));
        }

        public Entity GetEntity(params Specification<Entity>[] specifications)
        {
            return this.Find(specifications).FirstOrDefault();
        }

        public IEnumerable<Entity> GetEntities(params Specification<Entity>[] specifications)
        {
            return this.Find(specifications);
        }

        public IEnumerable<Entity> GetEntitiesPaged(string partitionKey, int pageIndex, int pageSize)
        {
            var results = this.Find(new ByPartitionKeySpecification(partitionKey));

            return results.Skip(pageIndex * pageSize).Take(pageSize);
        }

        public IEnumerable<Entity> Find(params Specification<Entity>[] specifications)
        {
            var predicate = Specification<Entity>.All(specifications);

            return predicate == null
                ? this.Client.Query<Entity>((string)null).ToArray()
                : this.Client.Query<Entity>(predicate).ToArray();
        }
    }
}
