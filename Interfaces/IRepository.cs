namespace AzurePatterns.Interfaces
{
    using System.Collections.Generic;

    using Azure.Data.Tables;

    public interface IRepository<TEntity> where TEntity : class, ITableEntity, new()
    {
        IEnumerable<TEntity> Find(params Specification<TEntity>[] specifications);

        void Save(TEntity item);
    }
}