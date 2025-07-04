#!/bin/bash

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Grafana Dashboard Generator Demo     ${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

echo -e "${GREEN}🚀 Step 1: Generating dashboard from sample API...${NC}"
make generate-sample
echo ""

echo -e "${GREEN}📊 Step 2: Validating generated dashboard...${NC}"
make validate
echo ""

echo -e "${GREEN}🐳 Step 3: Starting monitoring stack with sample API...${NC}"
make start-fresh
echo ""

echo -e "${YELLOW}⏳ Waiting for services to start up (30 seconds)...${NC}"
sleep 30

echo -e "${GREEN}🔍 Step 4: Checking service health...${NC}"

# Check Prometheus
if curl -s http://localhost:9090/-/ready > /dev/null 2>&1; then
    echo "✅ Prometheus is ready at http://localhost:9090"
else
    echo "❌ Prometheus is not ready"
fi

# Check Grafana
if curl -s http://localhost:3000/api/health > /dev/null 2>&1; then
    echo "✅ Grafana is ready at http://localhost:3000"
else
    echo "❌ Grafana is not ready"
fi

# Check Sample API
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "✅ Sample API is ready at http://localhost:8080"
else
    echo "❌ Sample API is not ready"
fi

echo ""
echo -e "${GREEN}🎯 Step 5: Testing sample API endpoints...${NC}"

# Test health endpoints
echo "Testing health endpoints:"
curl -s http://localhost:8080/api/inventory/v1/livez | jq '.status' 2>/dev/null || echo "Livez endpoint responded"
curl -s http://localhost:8080/api/inventory/v1/readyz | jq '.status' 2>/dev/null || echo "Readyz endpoint responded"

# Test auth endpoint
echo "Testing auth endpoint:"
curl -s -X POST http://localhost:8080/api/inventory/v1beta1/authz/check \
  -H "Content-Type: application/json" \
  -d '{"resource": "k8s-cluster", "action": "read", "subject": "user:demo@example.com"}' \
  | jq '.allowed' 2>/dev/null || echo "Auth endpoint responded"

echo ""
echo -e "${GREEN}📈 Step 6: Generating some traffic for metrics...${NC}"
for i in {1..10}; do
    curl -s http://localhost:8080/api/inventory/v1/livez > /dev/null &
    curl -s http://localhost:8080/api/inventory/v1/readyz > /dev/null &
    curl -s -X POST http://localhost:8080/api/inventory/v1beta1/resources/k8s-clusters \
      -H "Content-Type: application/json" \
      -d '{"name": "demo-cluster", "node_count": 3}' > /dev/null &
done
wait

echo "Generated traffic to create metrics data"
echo ""

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}           🎉 Demo Complete!            ${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${GREEN}🌐 Access the services:${NC}"
echo "  📊 Grafana Dashboard: http://localhost:3000 (admin/admin)"
echo "  📈 Prometheus:        http://localhost:9090"
echo "  🔧 Sample API:        http://localhost:8080"
echo "  📝 API Docs:          http://localhost:8080/health"
echo ""
echo -e "${GREEN}📋 Generated Files:${NC}"
echo "  📄 Dashboard JSON:    output_dashboard.json"
echo "  📄 Sample API Spec:   sample-openapi.yaml"
echo ""
echo -e "${GREEN}🛠️  Next Steps:${NC}"
echo "  1. Open Grafana at http://localhost:3000"
echo "  2. Login with admin/admin"
echo "  3. Navigate to the 'Sample API Dashboard'"
echo "  4. Watch the metrics populate as the API generates background traffic"
echo ""
echo -e "${YELLOW}💡 Pro Tips:${NC}"
echo "  • The sample API generates realistic traffic patterns"
echo "  • Try different time ranges in Grafana (last 5m, 15m, 1h)"
echo "  • Check the Prometheus targets at http://localhost:9090/targets"
echo "  • View raw metrics at http://localhost:8080/metrics"
echo ""
echo -e "${YELLOW}🧹 To clean up:${NC} make clean"
echo -e "${YELLOW}📖 For help:${NC}    make help" 