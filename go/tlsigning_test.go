package tlsigning

import (
	"os"
	"testing"

	"github.com/Truelayer/truelayer-signing/go/errors"
	"github.com/stretchr/testify/assert"
)

const (
	Kid = "45fc75cf-5649-4134-84b3-192c2c78e990"
)

func getTestKeys(assert *assert.Assertions) ([]byte, []byte) {
	privateKeyBytes, err := os.ReadFile("../test-resources/ec512-private.pem")
	assert.Nilf(err, "private key read failed: %v", err)
	publicKeyBytes, err := os.ReadFile("../test-resources/ec512-public.pem")
	assert.Nilf(err, "public key read failed: %v", err)
	return privateKeyBytes, publicKeyBytes
}

func getTlSignature(assert *assert.Assertions) string {
	signature, err := os.ReadFile("../test-resources/tl-signature.txt")
	assert.Nilf(err, "tl signature read failed: %v", err)
	return string(signature)
}

func getWebhookSignature(assert *assert.Assertions) string {
	signature, err := os.ReadFile("../test-resources/webhook-signature.txt")
	assert.Nilf(err, "webhook signature read failed: %v", err)
	return string(signature)
}

func getJwksJson(assert *assert.Assertions) []byte {
	jwks, err := os.ReadFile("../test-resources/jwks.json")
	assert.Nilf(err, "jwks json read failed: %v", err)
	return jwks
}

func TestVerifyV1StaticSignatureShouldFail(t *testing.T) {
	assert := assert.New(t)

	_, publicKeyBytes := getTestKeys(assert)

	tlSignature := getTlSignature(assert)

	body := []byte("{\"abc\":123}")

	err := VerifyWithPem(publicKeyBytes).
		Body(body).
		Verify(tlSignature)
	assert.NotNilf(err, "v1 signature verification should fail: %v", err)
	assert.ErrorAs(&errors.JwsError{}, &err, "error should be a JwsError")
}

func TestSignature(t *testing.T) {
	assert := assert.New(t)

	privateKeyBytes, publicKeyBytes := getTestKeys(assert)

	body := []byte("{\"currency\":\"GBP\",\"max_amount_in_minor\":5000000}")
	idempotencyKey := []byte("idemp-2076717c-9005-4811-a321-9e0787fa0382")
	path := "/merchant_accounts/a61acaef-ee05-4077-92f3-25543a11bd8d/sweeping"

	signature, err := SignWithPem(Kid, privateKeyBytes).
		Method("post").
		Path(path).
		Header("Idempotency-Key", idempotencyKey).
		Body(body).
		Sign()
	assert.Nilf(err, "signing failed: %v", err)

	err = VerifyWithPem(publicKeyBytes).
		Method("POST").
		Path(path).
		RequireHeader("Idempotency-Key").
		Header("X-Whatever-2", []byte("t2345d")).
		Header("Idempotency-Key", idempotencyKey).
		Body(body).
		Verify(signature)
	assert.Nilf(err, "signature verification should not fail: %v", err)
}

func TestVerifyStaticSignature(t *testing.T) {
	assert := assert.New(t)

	_, publicKeyBytes := getTestKeys(assert)

	body := []byte("{\"currency\":\"GBP\",\"max_amount_in_minor\":5000000,\"name\":\"Foo???\"}")
	idempotencyKey := []byte("idemp-2076717c-9005-4811-a321-9e0787fa0382")
	path := "/merchant_accounts/a61acaef-ee05-4077-92f3-25543a11bd8d/sweeping"
	tlSignature := getTlSignature(assert)

	err := VerifyWithPem(publicKeyBytes).
		Method("POST").
		Path(path).
		Header("X-Whatever-2", []byte("t2345d")).
		Header("Idempotency-Key", idempotencyKey).
		Body(body).
		Verify(tlSignature)
	assert.Nilf(err, "signature verification should not fail: %v", err)
}

func TestSignatureMethodMismatch(t *testing.T) {
	assert := assert.New(t)

	privateKeyBytes, publicKeyBytes := getTestKeys(assert)

	body := []byte("{\"currency\":\"GBP\",\"max_amount_in_minor\":5000000}")
	idempotencyKey := []byte("idemp-2076717c-9005-4811-a321-9e0787fa0382")
	path := "/merchant_accounts/a61acaef-ee05-4077-92f3-25543a11bd8d/sweeping"

	signature, err := SignWithPem(Kid, privateKeyBytes).
		Method("post").
		Path(path).
		Header("Idempotency-Key", idempotencyKey).
		Body(body).
		Sign()
	assert.Nilf(err, "signing failed: %v", err)

	err = VerifyWithPem(publicKeyBytes).
		Method("DELETE"). // different
		Path(path).
		Header("X-Whatever-2", []byte("aoitbeh")).
		Header("Idempotency-Key", idempotencyKey).
		Body(body).
		Verify(signature)
	assert.NotNilf(err, "signature verification should fail: %v", err)
	assert.ErrorAs(&errors.JwsError{}, &err, "error should be a JwsError")
}

func TestSignatureHeaderMismatch(t *testing.T) {
	assert := assert.New(t)

	privateKeyBytes, publicKeyBytes := getTestKeys(assert)

	body := []byte("{\"currency\":\"GBP\",\"max_amount_in_minor\":5000000}")
	idempotencyKey := []byte("idemp-2076717c-9005-4811-a321-9e0787fa0382")
	path := "/merchant_accounts/a61acaef-ee05-4077-92f3-25543a11bd8d/sweeping"

	signature, err := SignWithPem(Kid, privateKeyBytes).
		Method("post").
		Path(path).
		Header("Idempotency-Key", idempotencyKey).
		Body(body).
		Sign()
	assert.Nilf(err, "signing failed: %v", err)

	err = VerifyWithPem(publicKeyBytes).
		Method("post").
		Path(path).
		Header("X-Whatever-2", []byte("aoitbeh")).
		Header("Idempotency-Key", []byte("something-else")).
		Body(body).
		Verify(signature)
	assert.NotNilf(err, "signature verification should fail: %v", err)
	assert.ErrorAs(&errors.JwsError{}, &err, "error should be a JwsError")
}

func TestSignatureBodyMismatch(t *testing.T) {
	assert := assert.New(t)

	privateKeyBytes, publicKeyBytes := getTestKeys(assert)

	body := []byte("{\"currency\":\"GBP\",\"max_amount_in_minor\":5000000}")
	idempotencyKey := []byte("idemp-2076717c-9005-4811-a321-9e0787fa0382")
	path := "/merchant_accounts/a61acaef-ee05-4077-92f3-25543a11bd8d/sweeping"

	signature, err := SignWithPem(Kid, privateKeyBytes).
		Method("post").
		Path(path).
		Header("Idempotency-Key", idempotencyKey).
		Body(body).
		Sign()
	assert.Nilf(err, "signing failed: %v", err)

	err = VerifyWithPem(publicKeyBytes).
		Method("post").
		Path(path).
		Header("X-Whatever-2", []byte("aoitbeh")).
		Header("Idempotency-Key", idempotencyKey).
		Body([]byte("{\"max_amount_in_minor\":1234}")). // different
		Verify(signature)
	assert.NotNilf(err, "signature verification should fail: %v", err)
	assert.ErrorAs(&errors.JwsError{}, &err, "error should be a JwsError")
}

func TestSignatureMissingSignatureHeader(t *testing.T) {
	assert := assert.New(t)

	privateKeyBytes, publicKeyBytes := getTestKeys(assert)

	body := []byte("{\"currency\":\"GBP\",\"max_amount_in_minor\":5000000}")
	idempotencyKey := []byte("idemp-2076717c-9005-4811-a321-9e0787fa0382")
	path := "/merchant_accounts/a61acaef-ee05-4077-92f3-25543a11bd8d/sweeping"

	signature, err := SignWithPem(Kid, privateKeyBytes).
		Method("post").
		Path(path).
		Header("Idempotency-Key", idempotencyKey).
		Body(body).
		Sign()
	assert.Nilf(err, "signing failed: %v", err)

	err = VerifyWithPem(publicKeyBytes).
		Method("post").
		Path(path).
		Header("X-Whatever-2", []byte("aoitbeh")).
		// missing Idempotency-Key
		Body(body).
		Verify(signature)
	assert.NotNilf(err, "signature verification should fail: %v", err)
	assert.ErrorAs(&errors.JwsError{}, &err, "error should be a JwsError")
}

func TestRequiredHeaderMissingFromSignature(t *testing.T) {
	assert := assert.New(t)

	privateKeyBytes, publicKeyBytes := getTestKeys(assert)

	body := []byte("{\"currency\":\"GBP\",\"max_amount_in_minor\":5000000}")
	idempotencyKey := []byte("idemp-2076717c-9005-4811-a321-9e0787fa0382")
	path := "/merchant_accounts/a61acaef-ee05-4077-92f3-25543a11bd8d/sweeping"

	signature, err := SignWithPem(Kid, privateKeyBytes).
		Method("post").
		Path(path).
		Header("Idempotency-Key", idempotencyKey).
		Body(body).
		Sign()
	assert.Nilf(err, "signing failed: %v", err)

	err = VerifyWithPem(publicKeyBytes).
		Method("post").
		Path(path).
		RequireHeader("X-Required").
		Header("Idempotency-Key", idempotencyKey).
		Body(body).
		Verify(signature)
	assert.NotNilf(err, "signature verification should fail: %v", err)
	assert.ErrorAs(&errors.InvalidKeyError{}, &err, "error should be an InvalidKeyError")
}

func TestFlexibleHeaderCaseOrderVerify(t *testing.T) {
	assert := assert.New(t)

	privateKeyBytes, publicKeyBytes := getTestKeys(assert)

	body := []byte("{\"currency\":\"GBP\",\"max_amount_in_minor\":5000000}")
	idempotencyKey := []byte("idemp-2076717c-9005-4811-a321-9e0787fa0382")
	path := "/merchant_accounts/a61acaef-ee05-4077-92f3-25543a11bd8d/sweeping"

	signature, err := SignWithPem(Kid, privateKeyBytes).
		Method("post").
		Path(path).
		Header("Idempotency-Key", idempotencyKey).
		Header("X-Custom", []byte("123")).
		Body(body).
		Sign()
	assert.Nilf(err, "signing failed: %v", err)

	err = VerifyWithPem(publicKeyBytes).
		Method("POST").
		Path(path).
		Header("X-CUSTOM", []byte("123")).
		Header("idempotency-key", idempotencyKey).
		Body(body).
		Verify(signature)
	assert.Nilf(err, "signature verification should not fail: %v", err)
}

func TestEnforceDetached(t *testing.T) {
	assert := assert.New(t)

	_, publicKeyBytes := getTestKeys(assert)

	// signature for `/bar` but with a valid jws-body pre-attached
	tlSignature := "eyJhbGciOiJFUzUxMiIsImtpZCI6IjQ1ZmM3NWNmLTU2NDktNDEzNC04NGIzLTE5MmMyYzc4ZTk5MCIsInRsX3ZlcnNpb24iOiIyIiwidGxfaGVhZGVycyI6IiJ9.UE9TVCAvYmFyCnt9.ARLa7Q5b8k5CIhfy1qrS-IkNqCDeE-VFRDz7Lb0fXUMOi_Ktck-R7BHDMXFDzbI5TyaxIo5TGHZV_cs0fg96dlSxAERp3UaN2oCQHIE5gQ4m5uU3ee69XfwwU_RpEIMFypycxwq1HOf4LzTLXqP_CDT8DdyX8oTwYdUBd2d3D17Wd9UA"

	body := []byte("{}")
	path := "/foo"
	err := VerifyWithPem(publicKeyBytes).
		Method("post").
		Path(path).
		Body(body).
		Verify(tlSignature)
	assert.NotNilf(err, "signature verification should fail: %v", err)
	assert.ErrorAs(&errors.JwsError{}, &err, "error should be a JwsError")
}

func TestEnforceDetachedTrailingDots(t *testing.T) {
	assert := assert.New(t)

	_, publicKeyBytes := getTestKeys(assert)

	// signature for `/bar` but with a valid jws-body pre-attached
	tlSignature := "eyJhbGciOiJFUzUxMiIsImtpZCI6IjQ1ZmM3NWNmLTU2NDktNDEzNC04NGIzLTE5MmMyYzc4ZTk5MCIsInRsX3ZlcnNpb24iOiIyIiwidGxfaGVhZGVycyI6IiJ9.UE9TVCAvYmFyCnt9.ARLa7Q5b8k5CIhfy1qrS-IkNqCDeE-VFRDz7Lb0fXUMOi_Ktck-R7BHDMXFDzbI5TyaxIo5TGHZV_cs0fg96dlSxAERp3UaN2oCQHIE5gQ4m5uU3ee69XfwwU_RpEIMFypycxwq1HOf4LzTLXqP_CDT8DdyX8oTwYdUBd2d3D17Wd9UA...."

	body := []byte("{}")
	path := "/bar"
	err := VerifyWithPem(publicKeyBytes).
		Method("post").
		Path(path).
		Body(body).
		Verify(tlSignature)
	assert.NotNilf(err, "signature verification should fail: %v", err)
	assert.ErrorAs(&errors.JwsError{}, &err, "error should be a JwsError")
}

func TestSignatureNoHeaders(t *testing.T) {
	assert := assert.New(t)

	privateKeyBytes, publicKeyBytes := getTestKeys(assert)

	body := []byte("{\"currency\":\"GBP\",\"max_amount_in_minor\":5000000}")
	path := "/merchant_accounts/a61acaef-ee05-4077-92f3-25543a11bd8d/sweeping"

	signature, err := SignWithPem(Kid, privateKeyBytes).
		Method("post").
		Path(path).
		Body(body).
		Sign()
	assert.Nilf(err, "signing failed: %v", err)

	err = VerifyWithPem(publicKeyBytes).
		Method("post").
		Path(path).
		Header("X-Whatever", []byte("aoitbeh")).
		Body(body).
		Verify(signature)
	assert.Nilf(err, "signature verification should not fail: %v", err)
}

func TestVerifyWithoutMethodShouldFail(t *testing.T) {
	assert := assert.New(t)

	privateKeyBytes, publicKeyBytes := getTestKeys(assert)

	body := []byte("{\"currency\":\"GBP\",\"max_amount_in_minor\":5000000}")
	path := "/merchant_accounts/a61acaef-ee05-4077-92f3-25543a11bd8d/sweeping"

	signature, err := SignWithPem(Kid, privateKeyBytes).
		Path(path).
		Body(body).
		Sign()
	assert.Nilf(err, "signing failed: %v", err)

	err = VerifyWithPem(publicKeyBytes).
		Path(path).
		Body(body).
		Verify(signature)
	assert.NotNilf(err, "signature verification should fail: %v", err)
	assert.ErrorAs(&errors.JwsError{}, &err, "error should be a JwsError")
}

func TestJwsHeaderExtraction(t *testing.T) {
	assert := assert.New(t)

	webhookSignature := getWebhookSignature(assert)

	jwsHeader, _ := ExtractJwsHeader(webhookSignature)

	assert.Equal(jwsHeader.Alg, "ES512")
	assert.Equal(jwsHeader.Kid, Kid)
	assert.Equal(jwsHeader.TlVersion, "2")
	assert.Equal(jwsHeader.TlHeaders, "X-Tl-Webhook-Timestamp,Content-Type")
	assert.Equal(jwsHeader.Jku, "https://webhooks.truelayer.com/.well-known/jwks")
}

func TestVerifyWithJwks(t *testing.T) {
	assert := assert.New(t)

	webhookSignature := getWebhookSignature(assert)
	jwks := getJwksJson(assert)

	err := VerifyWithJwks(jwks).
		Method("POST").
		Path("/tl-webhook").
		Header("x-tl-webhook-timestamp", []byte("2021-11-29T11:42:55Z")).
		Header("content-type", []byte("application/json")).
		Body([]byte("{\"event_type\":\"example\",\"event_id\":\"18b2842b-a57b-4887-a0a6-d3c7c36f1020\"}")).
		Verify(webhookSignature)

	assert.Nilf(err, "signature verification should not fail: %v", err)

	err = VerifyWithJwks(jwks).
		Method("POST").
		Path("/tl-webhook").
		Header("x-tl-webhook-timestamp", []byte("2021-12-02T14:18:00Z")).
		Header("content-type", []byte("application/json")).
		Body([]byte("{\"event_type\":\"example\",\"event_id\":\"18b2842b-a57b-4887-a0a6-d3c7c36f1020\"}")).
		Verify(webhookSignature)

	assert.NotNilf(err, "signature verification should fail: %v", err)
	assert.ErrorAs(&errors.JwsError{}, &err, "error should be a JwsError")
}

func TestHeadersMethod(t *testing.T) {
	assert := assert.New(t)

	privateKeyBytes, publicKeyBytes := getTestKeys(assert)

	body := []byte("{\"currency\":\"GBP\",\"max_amount_in_minor\":5000000}")
	idempotencyKey := []byte("idemp-2076717c-9005-4811-a321-9e0787fa0382")
	path := "/merchant_accounts/a61acaef-ee05-4077-92f3-25543a11bd8d/sweeping"

	headers := make(map[string][]byte)
	headers["Idempotency-Key"] = idempotencyKey
	headers["X-Custom"] = []byte("123")
	signature, err := SignWithPem(Kid, privateKeyBytes).
		Method("post").
		Path(path).
		Headers(headers).
		Body(body).
		Sign()
	assert.Nilf(err, "signing failed: %v", err)

	err = VerifyWithPem(publicKeyBytes).
		Method("POST").
		Path(path).
		Headers(headers).
		Body(body).
		Verify(signature)
	assert.Nilf(err, "signature verification should not fail: %v", err)
}

func TestInvalidSignerPath(t *testing.T) {
	assert := assert.New(t)

	privateKeyBytes, _ := getTestKeys(assert)
	path := "https://example.com/the-path"

	_, err := SignWithPem(Kid, privateKeyBytes).
		Method("post").
		Path(path).
		Sign()
	assert.NotNilf(err, "signing should fail: %v", err)
	assert.ErrorAs(&errors.InvalidArgumentError{}, &err, "error should be an InvalidArgumentError")
}

func TestInvalidVerifierPath(t *testing.T) {
	assert := assert.New(t)

	_, publicKeyBytes := getTestKeys(assert)
	path := "https://example.com/the-path"
	signature := "signature"

	err := VerifyWithPem(publicKeyBytes).
		Method("POST").
		Path(path).
		Verify(signature)
	assert.NotNilf(err, "signature verification should fail: %v", err)
	assert.ErrorAs(&errors.InvalidArgumentError{}, &err, "error should be an InvalidArgumentError")
}

func TestSetJku(t *testing.T) {
	assert := assert.New(t)

	privateKeyBytes, _ := getTestKeys(assert)

	body := []byte("{\"event_type\":\"example\",\"event_id\":\"18b2842b-a57b-4887-a0a6-d3c7c36f1020\"}")
	headers := make(map[string][]byte)
	headers["X-Tl-Webhook-Timestamp"] = []byte("2021-11-29T11:42:55Z")
	headers["Content-Type"] = []byte("application/json")

	signature, err := SignWithPem(Kid, privateKeyBytes).
		Method("POST").
		Path("/tl-webhook").
		Headers(headers).
		Body(body).
		Jku("https://webhooks.truelayer.com/.well-known/jwks").
		Sign()
	assert.Nilf(err, "signing failed: %v", err)

	jwsHeader, err := ExtractJwsHeader(signature)
	assert.Nilf(err, "jws header extraction failed: %v", err)
	assert.Equal("https://webhooks.truelayer.com/.well-known/jwks", jwsHeader.Jku)
}

// Signing a path with a single trailing slash & trying to verify
// without that slash should still work. See #80.
func TestVerifyWithoutSignedTrailingSlash(t *testing.T) {
	assert := assert.New(t)

	privateKeyBytes, publicKeyBytes := getTestKeys(assert)

	body := []byte("{\"foo\":\"bar\"}")

	signature, err := SignWithPem(Kid, privateKeyBytes).
		Method("post").
		Path("/tl-webhook/").
		Body(body).
		Sign()
	assert.Nilf(err, "signing failed: %v", err)

	err = VerifyWithPem(publicKeyBytes).
		Method("POST").
		Path("/tl-webhook"). // missing trailing slash
		Body(body).
		Verify(signature)
	assert.Nilf(err, "signature verification should not fail: %v", err)
}

// Verify a path that matches except it has an additional trailing slash
// should still work. See #80.
func TestVerifyWithUnsignedTrailingSlash(t *testing.T) {
	assert := assert.New(t)

	privateKeyBytes, publicKeyBytes := getTestKeys(assert)

	body := []byte("{\"foo\":\"bar\"}")

	signature, err := SignWithPem(Kid, privateKeyBytes).
		Method("post").
		Path("/tl-webhook").
		Body(body).
		Sign()
	assert.Nilf(err, "signing failed: %v", err)

	err = VerifyWithPem(publicKeyBytes).
		Method("POST").
		Path("/tl-webhook/"). // additional trailing slash
		Body(body).
		Verify(signature)
	assert.Nilf(err, "signature verification should not fail: %v", err)
}
