namespace AzurePatterns.Tests
{
    using System.Linq;

    using AzurePatterns.Entities;
    using AzurePatterns.Repositories;

    using Xunit;

    public class SpecificationTests
    {
        [Fact]
        public void All_WithNoSpecifications_ReturnsNull()
        {
            Assert.Null(Specification<Entity>.All());
        }

        [Fact]
        public void All_CombinesSpecificationsWithAnd()
        {
            var predicate = Specification<Entity>.All(
                new ByPartitionKeySpecification("p"),
                new ByRowKeySpecification("r")).Compile();

            Assert.True(predicate(new Entity { PartitionKey = "p", RowKey = "r" }));
            Assert.False(predicate(new Entity { PartitionKey = "p", RowKey = "x" }));
            Assert.False(predicate(new Entity { PartitionKey = "x", RowKey = "r" }));
        }
    }
}
