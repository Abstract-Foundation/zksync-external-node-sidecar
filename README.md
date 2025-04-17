# ZKsync External Node Sidecar

Forked from [Stakewise ETH sidecar](https://github.com/stakewise/ethnode-sidecar/)

## Usage

To add liveness/readiness probes to an external node, add this container as a sidecar with probes to `/en/readiness` or `/en/liveness`

It will return OK if the node is fully synced (within 50 blocks of head) and the health check returns `ready`

### Example container config
```yaml
- name: sidecar
  image: ghcr.io/abstract-foundation/zksync-external-node-sidecar:latest
  imagePullPolicy: Always
  ports:
  - containerPort: 3000
    name: sidecar
    protocol: TCP
  # NOTE: Disable the liveness probe until node is synced, as it will fail if the node is not 
  # returning a synced state within 15 minutes of starting up.
  livenessProbe:
    httpGet:
      path: /en/liveness
      port: sidecar
    initialDelaySeconds: 900
    periodSeconds: 10
  readinessProbe:
    httpGet:
      path: /en/readiness
      port: sidecar
    initialDelaySeconds: 10
    periodSeconds: 10
```