package handlers

import (
	stdErrors "errors"
	"net/http"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/rarimo/evm-airdrop-svc/internal/data"
	"github.com/rarimo/evm-airdrop-svc/internal/service/api"
	"github.com/rarimo/evm-airdrop-svc/internal/service/api/models"
	"github.com/rarimo/evm-airdrop-svc/internal/service/api/requests"
	zk "github.com/rarimo/zkverifier-kit"
	"github.com/rarimo/zkverifier-kit/identity"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"gitlab.com/distributed_lab/logan/v3"
)

// Full list of the OpenSSL signature algorithms and hash-functions is provided here:
// https://www.openssl.org/docs/man1.1.1/man3/SSL_CTX_set1_sigalgs_list.html

func CreateAirdrop(w http.ResponseWriter, r *http.Request) {
	req, err := requests.NewCreateAirdrop(r)
	if err != nil {
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	nullifier := req.Data.Attributes.ZkProof.PubSignals[zk.Nullifier]

	airdrop, err := api.AirdropsQ(r).
		FilterByNullifier(nullifier).
		FilterByStatuses(data.TxStatusCompleted, data.TxStatusPending, data.TxStatusInProgress).
		Get()
	if err != nil {
		api.Log(r).WithError(err).Error("Failed to get airdrop by nullifier")
		ape.RenderErr(w, problems.InternalError())
		return
	}
	if airdrop != nil {
		ape.RenderErr(w, problems.Conflict())
		return
	}

	decodedAddress, err := hexutil.Decode(req.Data.Attributes.Address)
	if err != nil {
		api.Log(r).WithError(err).WithFields(logan.F{
			"address": req.Data.Attributes.Address,
		}).Error("Failed to decode hex ethereum address")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	err = api.Verifier(r).VerifyProof(req.Data.Attributes.ZkProof, zk.WithEventData(decodedAddress))
	if err != nil {
		if stdErrors.Is(err, identity.ErrContractCall) {
			api.Log(r).WithError(err).Error("Failed to verify proof")
			ape.RenderErr(w, problems.InternalError())
			return
		}

		api.Log(r).WithError(err).Info("Invalid proof")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	airdrop, err = api.AirdropsQ(r).Insert(data.Airdrop{
		Nullifier: nullifier,
		Address:   req.Data.Attributes.Address,
		Amount:    api.AirdropConfig(r).Amount.String(),
		Status:    data.TxStatusPending,
	})
	if err != nil {
		api.Log(r).WithError(err).Errorf("Failed to insert airdrop")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	w.WriteHeader(http.StatusCreated)
	ape.Render(w, models.NewAirdropResponse(*airdrop))
}
