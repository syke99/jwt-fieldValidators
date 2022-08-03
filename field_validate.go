package jwt

import (
	"errors"
	"sort"
	"time"
)

type JwtFieldValidator func(tokenClaims *Claims) error

func IssuerValidator(expectedIss string) JwtFieldValidator {
	return func(tokenClaims *Claims) error {
		tokenIss, ok := tokenClaims.String("iss")
		if !ok {
			return errors.New("jwt issuers is missing and is required")
		}

		if tokenIss != expectedIss {
			return errors.New("invalid issuer claim")
		}

		return nil
	}
}

func SubjectValidator(expectedSub string) JwtFieldValidator {
	return func(tokenClaims *Claims) error {
		tokenSub, ok := tokenClaims.String("sub")
		if !ok {
			return errors.New("missing subject claim")
		}

		if tokenSub != expectedSub {
			return errors.New("invalid subject claim")
		}

		return nil
	}
}

func AudiencesValidator(expectedAud []string) JwtFieldValidator {
	return func(tokenClaims *Claims) error {
		tokenAud := tokenClaims.Audiences
		if len(tokenAud) == 0 {
			return errors.New("missing audience claim")
		}

		if len(tokenAud) != len(expectedAud) {
			return errors.New("invalid audience claim")
		}

		sort.Strings(tokenAud)
		sort.Strings(expectedAud)

		for i := range tokenAud {
			if tokenAud[i] != expectedAud[i] {
				return errors.New("invalid audience claim")
			}
		}

		return nil
	}
}

func TimeFieldValidator(expectedTime time.Time) JwtFieldValidator {
	return func(tokenClaims *Claims) error {
		if ok := tokenClaims.Valid(expectedTime); !ok {
			return errors.New("token has expired")
		}

		return nil
	}
}

func IdValidator(expectedId string) JwtFieldValidator {
	return func(tokenClaims *Claims) error {
		tokenId, ok := tokenClaims.String("jti")
		if !ok {
			return errors.New("missing id claim")
		}

		if tokenId != expectedId {
			return errors.New("invalid id claim")
		}

		return nil
	}
}

func CustomFieldValidator(expectedValue, customField string) JwtFieldValidator {
	return func(tokenClaims *Claims) error {
		fieldValue, ok := tokenClaims.String(customField)
		if !ok {
			return errors.New("missing custom claim")
		}

		if fieldValue != expectedValue {
			return errors.New("invalid custom claim")
		}

		return nil
	}
}

func ValidateTokenFields(tokenClaims *Claims, validators ...JwtFieldValidator) error {
	var err error

	for _, validator := range validators {
		err = validator(tokenClaims)
	}

	return err
}
