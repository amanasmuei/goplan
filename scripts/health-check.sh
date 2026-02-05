#!/usr/bin/env bash
# ===========================================
# GoPlan Health Check Script
# Check service health and status
# ===========================================

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
ENVIRONMENT="${ENVIRONMENT:-staging}"
NAMESPACE=""
VERBOSE="${VERBOSE:-false}"
CHECK_TYPE="${CHECK_TYPE:-all}"
TIMEOUT="${TIMEOUT:-10}"

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[OK]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1"
}

# Print usage
usage() {
    cat <<EOF
Usage: $(basename "$0") [OPTIONS]

Check GoPlan service health and status.

OPTIONS:
    -e, --environment   Environment to check (staging|production) [default: staging]
    -n, --namespace     Kubernetes namespace [default: goplan-{environment}]
    -c, --check         Check type: all, http, k8s, db, redis [default: all]
    -t, --timeout       HTTP timeout in seconds [default: 10]
    -v, --verbose       Verbose output
    -h, --help          Show this help message

EXAMPLES:
    $(basename "$0") -e production                 # Check all health in production
    $(basename "$0") -e staging -c http            # Check HTTP endpoints only
    $(basename "$0") -e production -c k8s -v       # Check K8s status verbosely

EXIT CODES:
    0   All checks passed
    1   One or more checks failed
    2   Invalid arguments

EOF
    exit 0
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -e|--environment)
                ENVIRONMENT="$2"
                shift 2
                ;;
            -n|--namespace)
                NAMESPACE="$2"
                shift 2
                ;;
            -c|--check)
                CHECK_TYPE="$2"
                shift 2
                ;;
            -t|--timeout)
                TIMEOUT="$2"
                shift 2
                ;;
            -v|--verbose)
                VERBOSE="true"
                shift
                ;;
            -h|--help)
                usage
                ;;
            *)
                log_error "Unknown option: $1"
                usage
                ;;
        esac
    done

    # Validate environment
    if [[ "$ENVIRONMENT" != "staging" && "$ENVIRONMENT" != "production" ]]; then
        log_error "Invalid environment: $ENVIRONMENT. Must be 'staging' or 'production'."
        exit 2
    fi

    # Validate check type
    if [[ ! "$CHECK_TYPE" =~ ^(all|http|k8s|db|redis)$ ]]; then
        log_error "Invalid check type: $CHECK_TYPE"
        exit 2
    fi

    # Set default namespace if not provided
    if [[ -z "$NAMESPACE" ]]; then
        if [[ "$ENVIRONMENT" == "production" ]]; then
            NAMESPACE="goplan-prod"
        else
            NAMESPACE="goplan-staging"
        fi
    fi
}

# Set environment URLs
set_urls() {
    if [[ "$ENVIRONMENT" == "production" ]]; then
        API_HOST="https://api.goplan.io"
        FRONTEND_HOST="https://goplan.io"
    else
        API_HOST="https://api.staging.goplan.io"
        FRONTEND_HOST="https://staging.goplan.io"
    fi
}

# Track overall status
OVERALL_STATUS=0

mark_failed() {
    OVERALL_STATUS=1
}

# HTTP health checks
check_http() {
    log_info "Checking HTTP endpoints..."
    echo

    # Backend health
    echo -n "  Backend /health: "
    if response=$(curl -sf --max-time "$TIMEOUT" "$API_HOST/health" 2>&1); then
        log_success "Healthy"
        if [[ "$VERBOSE" == "true" ]]; then
            echo "    Response: $response"
        fi
    else
        log_error "Unhealthy"
        mark_failed
    fi

    # Backend readiness
    echo -n "  Backend /ready: "
    if response=$(curl -sf --max-time "$TIMEOUT" "$API_HOST/ready" 2>&1); then
        log_success "Ready"
        if [[ "$VERBOSE" == "true" ]]; then
            echo "    Response: $response"
        fi
    else
        log_error "Not Ready"
        mark_failed
    fi

    # Backend version
    echo -n "  Backend /version: "
    if response=$(curl -sf --max-time "$TIMEOUT" "$API_HOST/version" 2>&1); then
        version=$(echo "$response" | grep -o '"version":"[^"]*"' | cut -d'"' -f4)
        log_success "Version: $version"
        if [[ "$VERBOSE" == "true" ]]; then
            echo "    Response: $response"
        fi
    else
        log_error "Failed"
        mark_failed
    fi

    # Frontend health
    echo -n "  Frontend /api/health: "
    if response=$(curl -sf --max-time "$TIMEOUT" "$FRONTEND_HOST/api/health" 2>&1); then
        log_success "Healthy"
        if [[ "$VERBOSE" == "true" ]]; then
            echo "    Response: $response"
        fi
    else
        log_error "Unhealthy"
        mark_failed
    fi

    # Frontend page load
    echo -n "  Frontend / (page load): "
    if curl -sf --max-time "$TIMEOUT" -o /dev/null "$FRONTEND_HOST/" 2>&1; then
        log_success "OK"
    else
        log_error "Failed"
        mark_failed
    fi

    echo
}

# Kubernetes health checks
check_k8s() {
    log_info "Checking Kubernetes resources in namespace: $NAMESPACE"
    echo

    # Check if kubectl is available
    if ! command -v kubectl &> /dev/null; then
        log_warning "kubectl not available, skipping K8s checks"
        return
    fi

    # Check if we can connect
    if ! kubectl cluster-info &> /dev/null; then
        log_warning "Cannot connect to Kubernetes cluster, skipping K8s checks"
        return
    fi

    # Backend deployment
    echo -n "  Backend Deployment: "
    if kubectl get deployment goplan-backend -n "$NAMESPACE" &> /dev/null; then
        ready=$(kubectl get deployment goplan-backend -n "$NAMESPACE" -o jsonpath='{.status.readyReplicas}')
        desired=$(kubectl get deployment goplan-backend -n "$NAMESPACE" -o jsonpath='{.spec.replicas}')
        if [[ "$ready" == "$desired" && -n "$ready" ]]; then
            log_success "$ready/$desired ready"
        else
            log_error "${ready:-0}/$desired ready"
            mark_failed
        fi
    else
        log_error "Not found"
        mark_failed
    fi

    # Frontend deployment
    echo -n "  Frontend Deployment: "
    if kubectl get deployment goplan-frontend -n "$NAMESPACE" &> /dev/null; then
        ready=$(kubectl get deployment goplan-frontend -n "$NAMESPACE" -o jsonpath='{.status.readyReplicas}')
        desired=$(kubectl get deployment goplan-frontend -n "$NAMESPACE" -o jsonpath='{.spec.replicas}')
        if [[ "$ready" == "$desired" && -n "$ready" ]]; then
            log_success "$ready/$desired ready"
        else
            log_error "${ready:-0}/$desired ready"
            mark_failed
        fi
    else
        log_error "Not found"
        mark_failed
    fi

    # Backend pods
    echo "  Backend Pods:"
    kubectl get pods -n "$NAMESPACE" -l app.kubernetes.io/component=backend --no-headers 2>/dev/null | while read -r line; do
        pod_name=$(echo "$line" | awk '{print $1}')
        pod_status=$(echo "$line" | awk '{print $3}')
        pod_ready=$(echo "$line" | awk '{print $2}')
        if [[ "$pod_status" == "Running" ]]; then
            echo -e "    $pod_name: ${GREEN}$pod_status${NC} ($pod_ready)"
        else
            echo -e "    $pod_name: ${RED}$pod_status${NC} ($pod_ready)"
            mark_failed
        fi
    done

    # Frontend pods
    echo "  Frontend Pods:"
    kubectl get pods -n "$NAMESPACE" -l app.kubernetes.io/component=frontend --no-headers 2>/dev/null | while read -r line; do
        pod_name=$(echo "$line" | awk '{print $1}')
        pod_status=$(echo "$line" | awk '{print $3}')
        pod_ready=$(echo "$line" | awk '{print $2}')
        if [[ "$pod_status" == "Running" ]]; then
            echo -e "    $pod_name: ${GREEN}$pod_status${NC} ($pod_ready)"
        else
            echo -e "    $pod_name: ${RED}$pod_status${NC} ($pod_ready)"
            mark_failed
        fi
    done

    # HPA status
    if [[ "$VERBOSE" == "true" ]]; then
        echo "  HorizontalPodAutoscalers:"
        kubectl get hpa -n "$NAMESPACE" --no-headers 2>/dev/null | while read -r line; do
            echo "    $line"
        done
    fi

    # Recent events (if verbose)
    if [[ "$VERBOSE" == "true" ]]; then
        echo "  Recent Events:"
        kubectl get events -n "$NAMESPACE" --sort-by='.lastTimestamp' 2>/dev/null | tail -5 | while read -r line; do
            echo "    $line"
        done
    fi

    echo
}

# Database connectivity check
check_db() {
    log_info "Checking database connectivity..."
    echo

    # Use the /ready endpoint which checks database
    echo -n "  Database (via /ready endpoint): "
    if response=$(curl -sf --max-time "$TIMEOUT" "$API_HOST/ready" 2>&1); then
        db_status=$(echo "$response" | grep -o '"database":"[^"]*"' | cut -d'"' -f4)
        if [[ "$db_status" == "ok" ]]; then
            log_success "Connected"
        else
            log_error "Failed: $db_status"
            mark_failed
        fi
    else
        log_error "Cannot reach health endpoint"
        mark_failed
    fi

    echo
}

# Redis connectivity check
check_redis() {
    log_info "Checking Redis connectivity..."
    echo

    # This would require exposing a Redis-specific health endpoint
    # For now, we check if Redis pod is running in K8s
    echo -n "  Redis: "

    if ! command -v kubectl &> /dev/null; then
        log_warning "kubectl not available"
        return
    fi

    if kubectl get pods -n "$NAMESPACE" -l app.kubernetes.io/component=redis --no-headers 2>/dev/null | grep -q "Running"; then
        log_success "Running"
    else
        log_warning "Not deployed in cluster (may be using external Redis)"
    fi

    echo
}

# Print summary
print_summary() {
    echo "=========================================="
    echo -n "Overall Status: "
    if [[ $OVERALL_STATUS -eq 0 ]]; then
        log_success "All checks passed"
    else
        log_error "Some checks failed"
    fi
    echo "=========================================="
}

# Main function
main() {
    echo "=========================================="
    echo "  GoPlan Health Check"
    echo "=========================================="
    echo

    parse_args "$@"
    set_urls

    log_info "Environment: $ENVIRONMENT"
    log_info "Namespace: $NAMESPACE"
    log_info "API Host: $API_HOST"
    log_info "Frontend Host: $FRONTEND_HOST"
    echo

    case "$CHECK_TYPE" in
        all)
            check_http
            check_k8s
            check_db
            check_redis
            ;;
        http)
            check_http
            ;;
        k8s)
            check_k8s
            ;;
        db)
            check_db
            ;;
        redis)
            check_redis
            ;;
    esac

    print_summary
    exit $OVERALL_STATUS
}

# Run main function
main "$@"
