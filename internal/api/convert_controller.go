package api

import (
	"net/http"

	"github.com/LittleAksMax/bids-policy-service/internal/service"
	"github.com/LittleAksMax/bids-util/requests"
)

type ConvertController struct {
	service service.ConvertServiceInterface
}

func NewConvertController(service service.ConvertServiceInterface) *ConvertController {
	return &ConvertController{
		service: service,
	}
}

func (pc *ConvertController) ConvertTreeToScript(w http.ResponseWriter, r *http.Request) {
	convertReq := requests.GetRequestBody[ConvertTreeToScriptRequest](r)
	if convertReq == nil {
		requests.WriteJSON(w, http.StatusBadRequest, requests.APIResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	script, err := pc.service.TreeToScript(&convertReq.Program)
	if err != nil {
		requests.WriteJSON(w, http.StatusBadRequest, requests.APIResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	requests.WriteJSON(w, http.StatusOK, requests.APIResponse{
		Success: true,
		Data:    ConvertScriptToTreeRequest{Script: script},
	})
}

func (pc *ConvertController) ConvertScriptToTree(w http.ResponseWriter, r *http.Request) {
	convertReq := requests.GetRequestBody[ConvertScriptToTreeRequest](r)
	if convertReq == nil {
		requests.WriteJSON(w, http.StatusBadRequest, requests.APIResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	program := pc.service.ScriptToTree(convertReq.Script)
	if program == nil {
		requests.WriteJSON(w, http.StatusBadRequest, requests.APIResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	requests.WriteJSON(w, http.StatusOK, requests.APIResponse{
		Success: true,
		Data:    ConvertTreeToScriptRequest{Program: *program},
	})
}
