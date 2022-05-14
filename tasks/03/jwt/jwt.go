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


// To mock time in tests
var timeFunc = time.Now

type header struct {
        Algorithm       SignMethod      `json:"alg"`
        Type            string          `json:"typ"`
}

// our payload format has structure {'d': something...., 'exp' : time...}
type payload struct {
        Data            interface{}     `json:"d"`
        ExpTime         int64           `json:"exp,omitempty"`
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

func b64Json(data interface{}) (string, error) {
        js, err := json.Marshal(data)
        if err != nil {
                return "", err
        }
        return b64.RawURLEncoding.EncodeToString(js), nil
}

func encodedHeaderAssembly(configuration *config) (string) {
        // there will be no error
        jsonB64Header, _ := b64Json(header {
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
        jsonB64Payload, _ := b64Json(payload {
                Data:           data,
                ExpTime:        expTime,
        })
        return jsonB64Payload
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

        mac := hmac.New(hashMethod, configuration.Key)
        mac.Write(jwt.Bytes())
        jwtHMACbytes := mac.Sum(nil)
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


func Decode(token []byte, data interface{}, opts ...Option) error {
        return nil
}


