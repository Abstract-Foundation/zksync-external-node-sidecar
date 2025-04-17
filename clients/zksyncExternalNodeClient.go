package clients

import (
	"fmt"
	"log"
	"net/http"

	"github.com/abstract-foundation/zksync-external-node-sidecar/common/hexutil"
	"github.com/abstract-foundation/zksync-external-node-sidecar/config"
	"github.com/go-resty/resty/v2"
)

type zksyncExternalNodeClient struct {
	cfg        *config.Config
	jsonRpcUrl string
	healthUrl  string
	client     *resty.Client
}

func NewZksyncExternalNodeClient() *zksyncExternalNodeClient {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}
	var jsonRpcUrl string
	jsonRpcUrl = cfg.Client.Scheme + "://" + cfg.Client.Host + ":" + cfg.Client.RpcPort

	var healthUrl string
	healthUrl = cfg.Client.Scheme + "://" + cfg.Client.Host + ":" + cfg.Client.HealthPort + "/health"

	client := resty.New()
	return &zksyncExternalNodeClient{
		cfg:        cfg,
		jsonRpcUrl: jsonRpcUrl,
		healthUrl:  healthUrl,
		client:     client,
	}
}

// HealthCheck Returns OK if the node is fully
// synchronized and ready to receive traffic
func (e *zksyncExternalNodeClient) HealthCheck(w http.ResponseWriter, r *http.Request) {
	type jsonRpcResponse struct {
		Jsonrpc string      `json:"jsonrpc"`
		ID      int         `json:"id"`
		Result  interface{} `json:"result"`
	}

	type healthResponse struct {
		Status string `json:"status"`
	}

	var ethSyncing jsonRpcResponse

	_, err := e.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(`{"jsonrpc":"2.0","method":"eth_syncing","params":[],"id":1}`).
		SetResult(&ethSyncing).
		Post(e.jsonRpcUrl)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var health healthResponse

	_, err = e.client.R().
		SetResult(&health).
		Get(e.healthUrl)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if health.Status != "ready" {
		fmt.Println(fmt.Sprintf("Health check failed. Status: %s", health.Status))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch ethSyncing.Result.(type) {
	default:
		w.WriteHeader(http.StatusInternalServerError)
	case bool:
		ethSyncingResult := ethSyncing.Result
		if ethSyncingResult == false {
			fmt.Fprintf(w, "StatusOK. External node is healthy.")
		}
	case map[string]interface{}:
		ethSyncingResult := ethSyncing.Result.(map[string]interface{})
		if hBlock, ok := ethSyncingResult["highestBlock"]; ok {
			highestBlock, err := hexutil.DecodeUint64(fmt.Sprintf("%s", hBlock))
			if err != nil {
				fmt.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			currentBlock, err := hexutil.DecodeUint64(fmt.Sprintf("%s", ethSyncingResult["currentBlock"]))
			if err != nil {
				fmt.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if highestBlock-currentBlock > 50 {
				fmt.Println(fmt.Sprintf("highestBlock-currentBlock < 50, highestBlock: %d, currentBlock: %d", highestBlock, currentBlock))
				w.WriteHeader(http.StatusInternalServerError)
				return
			} else {
				fmt.Fprintf(w, "StatusOK. External node is healthy.")
			}
		} else {
			fmt.Fprintf(w, "StatusOK. External node is healthy.")
		}
	}

}
