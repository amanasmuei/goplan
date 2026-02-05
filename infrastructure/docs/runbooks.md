# GoPlan Runbooks

## Alert Response Procedures

### GoPlanAPIHighErrorRate

**Severity:** Critical
**Threshold:** Error rate > 5% for 5 minutes

#### Symptoms
- HTTP 5xx responses exceeding 5% of total requests
- Users reporting failures or errors in the application

#### Investigation Steps

1. **Check application logs**
   ```bash
   kubectl logs -n goplan-production -l app=goplan-api --tail=100
   ```

2. **Check recent deployments**
   ```bash
   kubectl rollout history deployment/prod-goplan-api -n goplan-production
   ```

3. **Verify database connectivity**
   ```bash
   kubectl exec -n goplan-production deployment/prod-goplan-api -- curl -s localhost:8080/health
   ```

4. **Check resource utilization**
   ```bash
   kubectl top pods -n goplan-production -l app=goplan-api
   ```

#### Resolution Steps

1. **If recent deployment caused issues:**
   ```bash
   kubectl rollout undo deployment/prod-goplan-api -n goplan-production
   ```

2. **If database connection issues:**
   - Verify PostgreSQL service is running
   - Check connection pool settings
   - Verify credentials in secrets

3. **If resource exhaustion:**
   - Scale up replicas temporarily
   - Review HPA settings

---

### GoPlanAPIHighLatency

**Severity:** Warning
**Threshold:** P95 latency > 1 second for 5 minutes

#### Symptoms
- Slow API responses
- User-reported delays
- Timeout errors in frontend

#### Investigation Steps

1. **Identify slow endpoints**
   ```bash
   kubectl logs -n goplan-production -l app=goplan-api | grep -E "duration=[0-9]+ms" | sort -t= -k2 -rn | head -20
   ```

2. **Check database query performance**
   - Review pg_stat_statements for slow queries
   - Check for missing indexes

3. **Check external dependencies**
   - OpenAI API response times
   - Vector database (pgvector) performance

#### Resolution Steps

1. **Short-term:**
   - Scale up API replicas
   - Enable query caching if available

2. **Long-term:**
   - Add database indexes
   - Optimize slow queries
   - Implement response caching

---

### GoPlanAPIPodNotReady

**Severity:** Warning
**Threshold:** Available replicas < desired replicas for 10 minutes

#### Symptoms
- Pods in CrashLoopBackOff or Pending state
- Reduced API capacity

#### Investigation Steps

1. **Check pod status**
   ```bash
   kubectl get pods -n goplan-production -l app=goplan-api
   ```

2. **Describe problematic pods**
   ```bash
   kubectl describe pod <pod-name> -n goplan-production
   ```

3. **Check events**
   ```bash
   kubectl get events -n goplan-production --sort-by='.lastTimestamp'
   ```

4. **Check resource availability**
   ```bash
   kubectl describe nodes | grep -A5 "Allocated resources"
   ```

#### Resolution Steps

1. **If OOMKilled:**
   - Increase memory limits
   - Check for memory leaks

2. **If image pull errors:**
   - Verify image exists in registry
   - Check image pull secrets

3. **If scheduling issues:**
   - Check node capacity
   - Review pod affinity rules

---

### GoPlanAPIHighMemoryUsage

**Severity:** Warning
**Threshold:** Memory usage > 85% of limit for 10 minutes

#### Symptoms
- High memory utilization in pods
- Potential for OOMKilled events

#### Investigation Steps

1. **Check current memory usage**
   ```bash
   kubectl top pods -n goplan-production -l app=goplan-api
   ```

2. **Profile memory usage**
   - Enable Go pprof endpoint if available
   - Check for goroutine leaks

3. **Review recent code changes**
   - Check for memory-intensive operations
   - Review caching implementations

#### Resolution Steps

1. **Immediate:**
   - Restart affected pods (rolling restart)
   ```bash
   kubectl rollout restart deployment/prod-goplan-api -n goplan-production
   ```

2. **Short-term:**
   - Increase memory limits
   - Scale out to distribute load

3. **Long-term:**
   - Profile and fix memory leaks
   - Optimize memory-intensive operations

---

### GoPlanDatabaseConnectionPoolExhausted

**Severity:** Critical
**Threshold:** Active connections >= 80% of max_connections for 5 minutes

#### Symptoms
- Database connection errors
- Slow or failing API requests
- "too many connections" errors

#### Investigation Steps

1. **Check active connections**
   ```sql
   SELECT count(*) FROM pg_stat_activity WHERE datname = 'goplan';
   ```

2. **Identify connection sources**
   ```sql
   SELECT application_name, count(*)
   FROM pg_stat_activity
   WHERE datname = 'goplan'
   GROUP BY application_name;
   ```

3. **Check for idle connections**
   ```sql
   SELECT pid, now() - pg_stat_activity.query_start AS duration, query, state
   FROM pg_stat_activity
   WHERE datname = 'goplan' AND state != 'idle'
   ORDER BY duration DESC;
   ```

#### Resolution Steps

1. **Immediate:**
   - Terminate idle connections if safe
   ```sql
   SELECT pg_terminate_backend(pid)
   FROM pg_stat_activity
   WHERE datname = 'goplan'
   AND state = 'idle'
   AND query_start < now() - interval '10 minutes';
   ```

2. **Short-term:**
   - Increase max_connections (requires restart)
   - Scale down API replicas temporarily

3. **Long-term:**
   - Implement connection pooling (PgBouncer)
   - Review connection pool settings in application
   - Optimize query patterns to reduce connection hold time

---

### GoPlanDatabaseSlowQueries

**Severity:** Warning
**Threshold:** Average query time > 500ms for 10 minutes

#### Symptoms
- Slow API responses
- Database CPU spikes
- Lock contention

#### Investigation Steps

1. **Identify slow queries**
   ```sql
   SELECT query, calls, mean_exec_time, total_exec_time
   FROM pg_stat_statements
   WHERE dbid = (SELECT oid FROM pg_database WHERE datname = 'goplan')
   ORDER BY mean_exec_time DESC
   LIMIT 10;
   ```

2. **Check for missing indexes**
   ```sql
   SELECT relname, seq_scan, seq_tup_read, idx_scan, idx_tup_fetch
   FROM pg_stat_user_tables
   WHERE seq_scan > idx_scan
   ORDER BY seq_tup_read DESC
   LIMIT 10;
   ```

3. **Analyze query plans**
   ```sql
   EXPLAIN ANALYZE <slow_query>;
   ```

#### Resolution Steps

1. **Add missing indexes**
   - Create indexes based on query patterns
   - Consider partial indexes for filtered queries

2. **Optimize queries**
   - Review and rewrite inefficient queries
   - Add pagination for large result sets

3. **Database maintenance**
   - Run VACUUM ANALYZE
   - Update table statistics

---

## Deployment Procedures

### Standard Deployment

1. **Pre-deployment checklist:**
   - All tests passing in CI
   - Database migrations reviewed
   - Rollback plan documented

2. **Deploy to staging:**
   ```bash
   kubectl kustomize infrastructure/kubernetes/overlays/staging | kubectl apply -f -
   ```

3. **Verify staging:**
   - Run smoke tests
   - Check metrics and logs

4. **Deploy to production:**
   ```bash
   kubectl kustomize infrastructure/kubernetes/overlays/production | kubectl apply -f -
   ```

5. **Monitor deployment:**
   ```bash
   kubectl rollout status deployment/prod-goplan-api -n goplan-production
   ```

### Rollback Procedure

1. **Identify the last working revision:**
   ```bash
   kubectl rollout history deployment/prod-goplan-api -n goplan-production
   ```

2. **Rollback to previous revision:**
   ```bash
   kubectl rollout undo deployment/prod-goplan-api -n goplan-production
   ```

3. **Rollback to specific revision:**
   ```bash
   kubectl rollout undo deployment/prod-goplan-api -n goplan-production --to-revision=<N>
   ```

4. **Verify rollback:**
   ```bash
   kubectl get pods -n goplan-production -l app=goplan-api
   kubectl logs -n goplan-production -l app=goplan-api --tail=50
   ```

---

## Escalation Contacts

| Level | Contact | Response Time |
|-------|---------|---------------|
| L1 | On-call Engineer | 15 minutes |
| L2 | Platform Team Lead | 30 minutes |
| L3 | Engineering Manager | 1 hour |

## Related Documentation

- [Architecture Overview](./architecture.md)
- [API Documentation](./api.md)
- [Database Schema](./database.md)
