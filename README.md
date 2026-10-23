# AI Software Engineering Agent (Zero-Cost Production Stack)

An autonomous AI Software Engineering Agent backend written in Go. It triages errors, searches code/logs/DB using local tools, generates root-cause analyses, and performs risk-based Human-in-the-Loop validation.

## 🚀 Quick Start (Run locally with Go)

```bash
# 1. Start Go server
go run cmd/server/main.go

# 2. Test the endpoint in another terminal
curl -X POST http://localhost:8080/api/triage \
  -H "Content-Type: application/json" \
  -d '{"issue": "The /api/orders endpoint is returning 500 errors when order contains more than 50 items."}'
```

## 🐳 Run with Docker Compose (Full Zero-Cost Infrastructure)

```bash
cd deployments
docker compose up -d --build

# Pull free local coding LLM (Ollama)
docker exec -it deployments-ollama-1 ollama run qwen2.5-coder:7b
```

## ☸️ Deploy to Local Kubernetes (Minikube / Kind)

```bash
minikube start
eval $(minikube docker-env)
docker build -t ai-agent-backend:latest .
kubectl apply -f deployments/k8s/deployment.yaml
```
