namespace AzurePatterns.Tests
{
    using System;
    using System.Linq;

    using AzurePatterns.Entities;
    using AzurePatterns.Repositories;
    using AzurePatterns.TableStorage;

    using Xunit;

    // Integration tests: run against Azurite (the Azure Storage emulator) using
    // UseDevelopmentStorage=true. They are skipped when Azurite is not running.
    public class EntityRepositoryTests
    {
        private const string ConnectionString = "UseDevelopmentStorage=true";

        [SkippableFact]
        public void SaveThenFind_RoundTripsEntity()
        {
            var context = Connect();
            var repository = new EntityRepository(context);
            var partition = "p-" + Guid.NewGuid().ToString("N");

            repository.Save(new Entity { PartitionKey = partition, RowKey = "r1" });
            repository.Save(new Entity { PartitionKey = partition, RowKey = "r2" });
            context.Commit();

            var found = repository.GetEntity(
                new ByPartitionKeySpecification(partition), new ByRowKeySpecification("r2"));

            Assert.NotNull(found);
            Assert.Equal("r2", found.RowKey);
            Assert.Equal(2, repository.GetEntities(new ByPartitionKeySpecification(partition)).Count());
        }

        [SkippableFact]
        public void Rollback_DiscardsPendingChanges()
        {
            var context = Connect();
            var repository = new EntityRepository(context);
            var partition = "p-" + Guid.NewGuid().ToString("N");

            repository.Save(new Entity { PartitionKey = partition, RowKey = "r1" });
            context.Rollback();
            context.Commit();

            Assert.Empty(repository.GetEntities(new ByPartitionKeySpecification(partition)));
        }

        [SkippableFact]
        public void Delete_RemovesEntity()
        {
            var context = Connect();
            var repository = new EntityRepository(context);
            var partition = "p-" + Guid.NewGuid().ToString("N");
            var entity = new Entity { PartitionKey = partition, RowKey = "r1" };

            repository.Save(entity);
            context.Commit();
            repository.Delete(entity);
            context.Commit();

            Assert.Empty(repository.GetEntities(new ByPartitionKeySpecification(partition)));
        }

        [SkippableFact]
        public void GetEntitiesPaged_HonoursPartitionKeyAndPaging()
        {
            var context = Connect();
            var repository = new EntityRepository(context);
            var partition = "p-" + Guid.NewGuid().ToString("N");

            for (var i = 0; i < 5; i++)
            {
                repository.Save(new Entity { PartitionKey = partition, RowKey = "r" + i });
            }

            repository.Save(new Entity { PartitionKey = "other-" + partition, RowKey = "r0" });
            context.Commit();

            Assert.Equal(2, repository.GetEntitiesPaged(partition, 0, 2).Count());
            Assert.Single(repository.GetEntitiesPaged(partition, 2, 2));
        }

        private static TableStorageContext Connect()
        {
            try
            {
                return new EntityRepositoryFixture().Create(ConnectionString);
            }
            catch (Exception exception)
            {
                Skip.If(true, "Azurite not available: " + exception.Message);
                throw;
            }
        }

        private sealed class EntityRepositoryFixture
        {
            public TableStorageContext Create(string connectionString)
            {
                var context = new TableStorageContext(connectionString);

                // Fail fast (and skip) when the emulator is not listening.
                foreach (var _ in context.Service.Query(maxPerPage: 1).Take(1))
                {
                }

                return context;
            }
        }
    }
}
