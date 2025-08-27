🚀 API Rate Limiter

📌 Overview

The API Rate Limiter is a distributed system that controls the rate of incoming requests to services deployed on Kubernetes.
It ensures fair usage, prevents abuse, and provides observability by exposing metrics to Prometheus for monitoring.

🛠️ Components

Rate Limiter Service (Custom Application)
Implements request throttling logic.

Exposes metrics at /metrics endpoint (Prometheus format).
Kubernetes Deployment & Service
Deploys the rate limiter as pods.
Exposes the service internally within the cluster.

Prometheus
Scrapes metrics from the Rate Limiter service.
Provides monitoring & visualization.

Minikube
Local Kubernetes cluster used to run the entire setup.

⚙️ How It Works

Client requests are sent to the Rate Limiter Service.

The service checks if the request exceeds the allowed rate:
 If allowed → forwards the request.
 If exceeded → returns HTTP 429 (Too Many Requests).

The service exposes metrics (e.g., allowed requests, denied requests) at /metrics.
Prometheus scrapes these metrics from the service using Kubernetes DNS discovery.
Metrics can be visualized in Prometheus UI (and extended to Grafana).

🖥️ Steps to Run
1. Start Minikube
minikube start

2. Deploy Rate Limiter Service
kubectl apply -f rate-limiter-deployment.yaml
kubectl apply -f rate-limiter-service.yaml

3. Deploy Prometheus
kubectl apply -f prometheus-deployment.yaml
kubectl apply -f prometheus-service.yaml

4. Verify Pods & Services
kubectl get pods -n monitoring
kubectl get svc -n monitoring

5. Access Prometheus Dashboard
minikube service prometheus -n monitoring

6. Test Rate Limiter

Get Minikube IP:
minikube ip
Then test with:
curl http://<MINIKUBE_IP>:<NODEPORT>/your-endpoint

7. Check Metrics in Prometheus

Go to Prometheus UI → Targets → You should see cloud-rate-limiter status as UP.
