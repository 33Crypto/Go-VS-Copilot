namespace AzurePatterns
{
    using System;
    using System.Linq;
    using System.Linq.Expressions;

    public abstract class Specification<TEntity>
    {
        public abstract Expression<Func<TEntity, bool>> Predicate { get; }

        // Combines specifications into a single AND predicate; null when there are none.
        public static Expression<Func<TEntity, bool>> All(params Specification<TEntity>[] specifications)
        {
            if (specifications == null || specifications.Length == 0)
            {
                return null;
            }

            var parameter = Expression.Parameter(typeof(TEntity), "e");
            Expression body = null;

            foreach (var predicate in specifications.Select(s => s.Predicate))
            {
                var next = new ReplaceParameter(predicate.Parameters[0], parameter).Visit(predicate.Body);
                body = body == null ? next : Expression.AndAlso(body, next);
            }

            return Expression.Lambda<Func<TEntity, bool>>(body, parameter);
        }

        private sealed class ReplaceParameter : ExpressionVisitor
        {
            private readonly ParameterExpression from;
            private readonly ParameterExpression to;

            public ReplaceParameter(ParameterExpression from, ParameterExpression to)
            {
                this.from = from;
                this.to = to;
            }

            protected override Expression VisitParameter(ParameterExpression node)
            {
                return node == this.from ? this.to : base.VisitParameter(node);
            }
        }
    }
}
