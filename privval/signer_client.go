package privval

import (
	"fmt"
	"time"

	pvproto "github.com/cometbft/cometbft/api/cometbft/privval/v2"
	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v2"
	"github.com/cometbft/cometbft/v2/crypto"
	cryptoenc "github.com/cometbft/cometbft/v2/crypto/encoding"
	"github.com/cometbft/cometbft/v2/types"
	cmterrors "github.com/cometbft/cometbft/v2/types/errors"
)

// SignerClient implements PrivValidator.
// Handles remote validator connections that provide signing services.
type SignerClient struct {
	endpoint *SignerListenerEndpoint
	chainID  string
}

var _ types.PrivValidator = (*SignerClient)(nil)

// NewSignerClient returns an instance of SignerClient.
// it will start the endpoint (if not already started).
func NewSignerClient(endpoint *SignerListenerEndpoint, chainID string) (*SignerClient, error) {
	if !endpoint.IsRunning() {
		if err := endpoint.Start(); err != nil {
			return nil, fmt.Errorf("failed to start listener endpoint: %w", err)
		}
	}

	return &SignerClient{endpoint: endpoint, chainID: chainID}, nil
}

// Close closes the underlying connection.
func (sc *SignerClient) Close() error {
	return sc.endpoint.Close()
}

// IsConnected indicates with the signer is connected to a remote signing service.
func (sc *SignerClient) IsConnected() bool {
	return sc.endpoint.IsConnected()
}

// WaitForConnection waits maxWait for a connection or returns a timeout error.
func (sc *SignerClient) WaitForConnection(maxWait time.Duration) error {
	return sc.endpoint.WaitForConnection(maxWait)
}

// --------------------------------------------------------
// Implement PrivValidator

// Ping sends a ping request to the remote signer.
func (sc *SignerClient) Ping() error {
	response, err := sc.endpoint.SendRequest(mustWrapMsg(&pvproto.PingRequest{}))
	if err != nil {
		sc.endpoint.Logger.Error("SignerClient::Ping", "err", err)
		return nil
	}

	pb := response.GetPingResponse()
	if pb == nil {
		return err
	}

	return nil
}

// GetPubKey retrieves a public key from a remote signer
// returns an error if client is not able to provide the key.
func (sc *SignerClient) GetPubKey() (crypto.PubKey, error) {
	response, err := sc.endpoint.SendRequest(mustWrapMsg(&pvproto.PubKeyRequest{ChainId: sc.chainID}))
	if err != nil {
		return nil, fmt.Errorf("send: %w", err)
	}

	resp := response.GetPubKeyResponse()
	if resp == nil {
		return nil, cmterrors.ErrRequiredField{Field: "response"}
	}
	if resp.Error != nil {
		return nil, &RemoteSignerError{Code: int(resp.Error.Code), Description: resp.Error.Description}
	}

	pk, err := cryptoenc.PubKeyFromTypeAndBytes(resp.PubKeyType, resp.PubKeyBytes)
	if err != nil {
		return nil, err
	}

	return pk, nil
}

// SignVote requests a remote signer to sign a vote.
func (sc *SignerClient) SignVote(chainID string, vote *cmtproto.Vote, signExtension bool) error {
	response, err := sc.endpoint.SendRequest(mustWrapMsg(&pvproto.SignVoteRequest{Vote: vote, ChainId: chainID, SkipExtensionSigning: !signExtension}))
	if err != nil {
		return err
	}

	resp := response.GetSignedVoteResponse()
	if resp == nil {
		return cmterrors.ErrRequiredField{Field: "response"}
	}
	if resp.Error != nil {
		return &RemoteSignerError{Code: int(resp.Error.Code), Description: resp.Error.Description}
	}

	*vote = resp.Vote

	return nil
}

// SignProposal requests a remote signer to sign a proposal.
func (sc *SignerClient) SignProposal(chainID string, proposal *cmtproto.Proposal) error {
	response, err := sc.endpoint.SendRequest(mustWrapMsg(
		&pvproto.SignProposalRequest{Proposal: proposal, ChainId: chainID},
	))
	if err != nil {
		return err
	}

	resp := response.GetSignedProposalResponse()
	if resp == nil {
		return cmterrors.ErrRequiredField{Field: "response"}
	}
	if resp.Error != nil {
		return &RemoteSignerError{Code: int(resp.Error.Code), Description: resp.Error.Description}
	}

	*proposal = resp.Proposal

	return nil
}

// SignBytes requests a remote signer to sign bytes.
func (sc *SignerClient) SignBytes(bytes []byte) ([]byte, error) {
	response, err := sc.endpoint.SendRequest(mustWrapMsg(&pvproto.SignBytesRequest{Value: bytes}))
	if err != nil {
		return nil, err
	}

	resp := response.GetSignBytesResponse()
	if resp == nil {
		return nil, cmterrors.ErrRequiredField{Field: "response"}
	}
	if resp.Error != nil {
		return nil, &RemoteSignerError{Code: int(resp.Error.Code), Description: resp.Error.Description}
	}

	return resp.Signature, nil
}
