package jwt

import (
        "errors"
        "time"
        "hash"
        "crypto/hmac"
        "crypto/sha256"
        "crypto/sha512"
        "bytes"
        "encoding/json"
        b64 "encoding/base64"
)


type header struct {
        Algorithm       SignMethod      `json:"alg"`
        Type            string          `json:"typ"`
}

// our payload format has structure {'d': something...., 'exp' : time...}
type payload struct {
        Data            interface{}	`json:"d"`
        ExpTime         int64           `json:"exp,omitempty"`
}

// special for decoding
type payloadDecoded struct {
        Data            json.RawMessage	`json:"d"`	// output
        ExpTime         int64           `json:"exp,omitempty"`
}

// special for decoding
type jwtParts struct {
        Header          header
        Payload		payloadDecoded
        Signature       []byte
}

type SignMethod string

const (
        HS256 SignMethod = "HS256"
        HS512 SignMethod = "HS512"
)

// convinient for code extension
var HASHSIGNFUNCTION = map[SignMethod] func() (hash.Hash) {
        HS256: sha256.New,
        HS512: sha512.New,
}

var (
        ErrInvalidSignMethod            = errors.New("invalid sign method")
        ErrSignatureInvalid             = errors.New("signature invalid")
        ErrTokenExpired                 = errors.New("token expired")
        ErrSignMethodMismatched         = errors.New("sign method mismatched")
        ErrConfigurationMalformed       = errors.New("configuration malformed")
        ErrInvalidToken                 = errors.New("invalid token")
        ErrInternalError                = errors.New("internal error")
)

func assemblyJWTConfig(opts []Option) (*config) {
        configuration := new(config)
        for _, option := range opts {
                option(configuration)
        }
        return configuration
}

func jwtConfigParse(configuration *config) (error) {
        if ((configuration.Expires != nil) && 
           ((configuration.TTL != nil) || configuration.Expires.Before(timeFunc()))) {
                return ErrConfigurationMalformed
        }
        return nil
}

func b64JsonEncode(data interface{}) (string, error) {
        js, err := json.Marshal(data)
        if err != nil {
                return "", err
        }
        return b64.RawURLEncoding.EncodeToString(js), nil
}

func b64JsonDecode(data []byte, dstData interface{}) (error) {
	decodedB64Data := make([]byte, b64.RawURLEncoding.DecodedLen(len(data)))
	_, err := b64.RawURLEncoding.Decode(decodedB64Data, data)
        if err != nil {
                return ErrInvalidToken
        }
        err = json.Unmarshal(decodedB64Data, dstData)
        if err != nil {
                return ErrInvalidToken
        }
        return nil
}

func encodedHeaderAssembly(configuration *config) (string) {
        // there will be no error
        jsonB64Header, _ := b64JsonEncode(header {
                Algorithm:      configuration.SignMethod,
                Type:           "JWT",
        })
        return jsonB64Header
}

func encodedPayloadAssembly(configuration *config, data interface{}) (string) {
        var expTime int64
        if configuration.Expires != nil {
                expTime = configuration.Expires.Unix()
        }
        if configuration.TTL != nil {
                expTime = timeFunc().Add(*configuration.TTL).Unix()
        }

        // else expTime will be 0 by default
        // there will be no error
        jsonB64Payload, _ := b64JsonEncode(payload {
                Data:           data,
                ExpTime:        expTime,
        })
        return jsonB64Payload
}

func hmacSignature(hashMethod func() (hash.Hash), key []byte, data []byte) []byte {
        mac := hmac.New(hashMethod, key)
        mac.Write(data)
        return mac.Sum(nil)
}

func jwtAssembly(configuration *config, jwtB64Header string, jwtB64Payload string) ([]byte, error) {
        var jwt bytes.Buffer
        jwt.WriteString(jwtB64Header)
        jwt.WriteString(".")
        jwt.WriteString(jwtB64Payload)

        hashMethod, ok := HASHSIGNFUNCTION[configuration.SignMethod]
        if !ok {
                return nil, ErrInvalidSignMethod
        }

        jwtHMACbytes := hmacSignature(hashMethod, configuration.Key, jwt.Bytes())
        jwtSign := b64.RawURLEncoding.EncodeToString(jwtHMACbytes)

        jwt.WriteString(".")
        jwt.WriteString(jwtSign)
        return jwt.Bytes(), nil
}

func Encode(data interface{}, opts ...Option) ([]byte, error) {
        jwtConfiguration := assemblyJWTConfig(opts)
        err := jwtConfigParse(jwtConfiguration)
        if err != nil {
                return nil, err
        }
        jwtHeader := encodedHeaderAssembly(jwtConfiguration)
        jwtPayload := encodedPayloadAssembly(jwtConfiguration, data)

        jwt, err := jwtAssembly(jwtConfiguration, jwtHeader, jwtPayload)
        if err != nil {
                return nil, err
        }
        return jwt, nil
}

func jwtB64DecodeParts(splittedParts [][]byte) (*jwtParts, error) {
        jwtparts := new(jwtParts)

        // check header
	err := b64JsonDecode(splittedParts[0], &jwtparts.Header)
        if err != nil {
                return nil, ErrInvalidToken
        }

        // check payload
        err = b64JsonDecode(splittedParts[1], &jwtparts.Payload)
        if err != nil {
                return nil, ErrInvalidToken
        }

        // decode hmac
	jwtparts.Signature = make([]byte, b64.RawURLEncoding.DecodedLen(len(splittedParts[2])))
	_, err = b64.RawURLEncoding.Decode(jwtparts.Signature, splittedParts[2])
	if err != nil {
                return nil, ErrSignatureInvalid
        }
        return jwtparts, nil
}

func jwtTokenSplitDecode(token []byte) (*jwtParts, error) {
        splittedParts := bytes.Split(token, []byte("."))
        // incorrent amount of partitions
        if len(splittedParts) != 3 {
                return nil, ErrInvalidToken
        }
        return jwtB64DecodeParts(splittedParts)
}

// true -> already expired
func jwtIsAlreadyExpired(jwt *jwtParts) (error) {
	// first - default variable when ExpTime is not located in JWT
        if jwt.Payload.ExpTime != 0 && timeFunc().After(time.Unix(jwt.Payload.ExpTime, 0)) {
		return ErrTokenExpired
	}
	return nil
}

func jwtCheckHeader(configuration *config, jwt *jwtParts) (error) {
        if jwt.Header.Type != "JWT" {
                return ErrInvalidToken
        }

        if jwt.Header.Algorithm != configuration.SignMethod {
                return ErrSignMethodMismatched
        }
        _, ok := HASHSIGNFUNCTION[jwt.Header.Algorithm]
        if !ok {
                return ErrInvalidSignMethod
        }
        return nil
}

// function takes slice to avoid redundant copying
func jwtSignatureValidation(configuration *config, expectedSignature []byte, headerAndPayload []byte) (error) {
        hashMethod, _ := HASHSIGNFUNCTION[configuration.SignMethod]
        if bytes.Compare(expectedSignature, hmacSignature(hashMethod, configuration.Key, headerAndPayload)) != 0 {
                return ErrSignatureInvalid
	}
	return nil
}


func Decode(token []byte, data interface{}, opts ...Option) error {
        jwtConfiguration := assemblyJWTConfig(opts)
        jwt, err := jwtTokenSplitDecode(token)
        if err != nil {
                return err
        }

        // check signature methods in header
        err = jwtCheckHeader(jwtConfiguration, jwt)
        if err != nil {
                return err
        }

	// work with token []byte to avoid redundant copying
	// validate signature
	headerAndPayload := token[:bytes.LastIndex(token, []byte("."))]
	err = jwtSignatureValidation(jwtConfiguration, jwt.Signature, headerAndPayload)
	if err != nil {
		return err
	}

        // check expire time
        err = jwtIsAlreadyExpired(jwt)
	if err != nil {
                return err
        }

	// extract data
	// Для анмаршалинга нужна "схема Json", то есть просто так jsonned interface{} -> interface{}
	// вроде бы нельзя сконвертировать. Поэтому я и заменил Payload.Data -> json.RawMessage
	err = json.Unmarshal(jwt.Payload.Data, data)
	if err != nil {
		return ErrInvalidToken
	}
	return nil
}


// To mock time in tests
var timeFunc = time.Now
