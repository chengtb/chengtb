// Package main demonstrates how to call the SUBMAIL SMS Send API.
// API Documentation: https://en.mysubmail.com/documents/3UQA3
//
// This demo shows two authentication modes:
//  1. Normal mode  - pass APPID and APPKEY directly as the signature (for testing only)
//  2. MD5 mode     - generate an MD5 hash signature with a server-side timestamp (recommended for production)
package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

// API endpoints
const (
	SMSSendAPI   = "https://api.mysubmail.com/message/send"
	TimestampAPI = "https://api.mysubmail.com/service/timestamp"
)

// Credentials — replace with your actual APPID and APPKEY from https://www.mysubmail.com
const (
	APPID  = "your_appid"
	APPKEY = "your_appkey"
)

// TimestampResponse is the JSON response returned by the SUBMAIL timestamp service.
type TimestampResponse struct {
	Timestamp float64 `json:"timestamp"`
}

// getTimestamp fetches the current server-side timestamp from SUBMAIL.
// Using the server timestamp (rather than the local clock) avoids clock-skew errors.
func getTimestamp() (string, error) {
	resp, err := http.Get(TimestampAPI)
	if err != nil {
		return "", fmt.Errorf("fetching timestamp: %w", err)
	}
	defer resp.Body.Close()

	var ts TimestampResponse
	if err := json.NewDecoder(resp.Body).Decode(&ts); err != nil {
		return "", fmt.Errorf("decoding timestamp response: %w", err)
	}
	return strconv.FormatFloat(ts.Timestamp, 'f', 0, 64), nil
}

// md5Hash returns the lowercase hexadecimal MD5 digest of s.
func md5Hash(s string) string {
	h := md5.New()
	io.WriteString(h, s) //nolint:errcheck // md5.Write never returns an error
	return hex.EncodeToString(h.Sum(nil))
}

// buildMD5Signature constructs the SUBMAIL sign_version=2 / sign_type=md5 signature.
//
// The algorithm (from the official SUBMAIL Go SDK) is:
//
//	APPID + APPKEY + sorted_params_string + APPID + APPKEY
//
// where sorted_params_string is the alphabetically sorted key=value pairs of ALL
// signed request fields (appid, sign_type, sign_version, timestamp, to) joined with
// "&".  The "content" / "vars" payload fields must be added to the request AFTER
// the signature is calculated, not before.
func buildMD5Signature(params map[string]string) string {
	// Collect all keys currently in the map (content is not yet present).
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
	}

	raw := APPID + APPKEY + strings.Join(parts, "&") + APPID + APPKEY
	return md5Hash(raw)
}

// postForm sends an application/x-www-form-urlencoded POST request to rawURL with
// the given fields and returns the response body as a string.
// This matches the encoding used by the official SUBMAIL Go SDK.
func postForm(rawURL string, fields map[string]string) (string, error) {
	form := url.Values{}
	for key, val := range fields {
		form.Set(key, val)
	}

	resp, err := http.Post(rawURL, "application/x-www-form-urlencoded;charset=utf-8", strings.NewReader(form.Encode())) //nolint:noctx
	if err != nil {
		return "", fmt.Errorf("POST %s: %w", rawURL, err)
	}
	defer resp.Body.Close()

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}
	return string(result), nil
}

// sendSMSNormal sends an SMS using the "normal" (non-encrypted) authentication mode.
// In this mode the APPKEY is sent directly as the "signature" field.
// This is the simplest mode but is only recommended for development / testing.
func sendSMSNormal(to, content string) error {
	fmt.Println("=== Normal Mode ===")
	params := map[string]string{
		"appid":     APPID,
		"signature": APPKEY, // plain-text APPKEY used as signature
		"to":        to,
		"content":   content,
	}

	result, err := postForm(SMSSendAPI, params)
	if err != nil {
		return err
	}
	fmt.Println("Response:", result)
	return nil
}

// sendSMSMD5 sends an SMS using the MD5-signed (sign_version=2) authentication mode.
// A server-issued timestamp is fetched first so that the signature is time-bound,
// making it significantly harder to replay intercepted requests.
//
// Important: "content" is intentionally added to the params map AFTER the signature
// is computed — this matches the official SUBMAIL SDK behaviour where the payload
// fields do not participate in the signature calculation.
func sendSMSMD5(to, content string) error {
	fmt.Println("=== MD5 Signature Mode ===")

	timestamp, err := getTimestamp()
	if err != nil {
		return fmt.Errorf("getting timestamp: %w", err)
	}

	// Build the params that participate in the signature.
	params := map[string]string{
		"appid":        APPID,
		"to":           to,
		"timestamp":    timestamp,
		"sign_type":    "md5",
		"sign_version": "2",
	}
	params["signature"] = buildMD5Signature(params)

	// Add payload fields only after the signature has been computed.
	params["content"] = content

	result, err := postForm(SMSSendAPI, params)
	if err != nil {
		return err
	}
	fmt.Println("Response:", result)
	return nil
}

func main() {
	// Replace these values with real ones before running.
	to := "138xxxxxxxx"                         // recipient mobile number
	content := "【SubMail】This is a test SMS." // message content (must include your registered SMS signature)

	// --- Normal mode ---------------------------------------------------------
	if err := sendSMSNormal(to, content); err != nil {
		fmt.Println("Normal mode error:", err)
	}

	// --- MD5 signature mode --------------------------------------------------
	if err := sendSMSMD5(to, content); err != nil {
		fmt.Println("MD5 mode error:", err)
	}
}
