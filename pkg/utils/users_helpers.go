package utils

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
)

var ErrInvalidUpdatedAt = errors.InvalidArgumentErrorf("invalid updated at timestamp")

func ValidatePassword(password string) error {
	if strings.TrimSpace(password) != password {
		return errors.InvalidArgumentErrorf("invalid password; starts or ends with a whitespace")
	}
	if password == "" {
		return errors.InvalidArgumentErrorf("invalid password; empty")
	}
	var (
		num, sym bool
		tot      uint8
	)
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			tot++
		case unicode.IsLower(char):
			tot++
		case unicode.IsNumber(char):
			num = true
			tot++
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			sym = true
			tot++
		default:
			return errors.InvalidArgumentErrorf("invalid password; contains invalid characters")
		}
	}
	err := errors.InvalidArgumentErrorf("invalid password; must contain a number, normal symbol and be at least 8 chars long")

	if !num || tot < 8 || !sym {
		return err
	}
	return nil
}

func ValidateEmail(email string) error {
	_, err := mail.ParseAddress(email)
	if err != nil && email != "" {
		return err
	}
	return nil
}

func ValidatePhone(phone string) error {
	re := regexp.MustCompile(`^(?:(?:\(?(?:00|\+)([1-4]\d\d|[1-9]\d?)\)?)?[\-\.\ \\\/]?)?((?:\(?\d{1,}\)?[\-\.\ \\\/]?){0,})(?:[\-\.\ \\\/]?(?:#|ext\.?|extension|x)[\-\.\ \\\/]?(\d+))?$`)
	if !re.MatchString(phone) && phone != "" {
		return errors.InvalidArgumentErrorf("invalid phone number")
	}
	return nil
}

func GetIdentifierFromIdentifier(userID uuid.UUID) (string, string) {
	return "id", userID.String()
}

func GetExponentialBackoff(backoff float64, factor float64) time.Duration {
	min := 100 * time.Millisecond
	max := 30 * time.Second
	minf := float64(min)
	durf := minf * math.Pow(factor, backoff)
	durf = rand.Float64()*(durf-minf) + minf
	dur := time.Duration(durf)
	// keep within bounds
	if dur < min {
		return min
	} else if dur > max {
		return max
	} else {
		return dur
	}
}

var ErrFileIsTooLarge = errors.InvalidArgumentErrorf("file exceeds maximum file size")

type DataReceive func() ([]byte, error)

func (d DataReceive) Receive() ([]byte, error) {
	return d()
}

func GetData(stream interface {
	Receive() ([]byte, error)
}, maxSize int,
) ([]byte, error) {
	size := 0
	data := bytes.Buffer{}
	defer data.Reset()
	for {
		req, err := stream.Receive()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		size += len(req)
		_, err = data.Write(req)
		if err != nil {
			return nil, err
		}
		if size > maxSize {
			return nil, ErrFileIsTooLarge
		}
	}
	return data.Bytes(), nil
}

func UserName(u *user.User) string {
	if n := u.GetProfile().GetFirstName(); n != "" {
		if l := u.GetProfile().GetLastName(); l != "" {
			n = fmt.Sprint(n, " ", l)
		}
		return n
	}
	if n := u.GetUserInfo().GetUsername(); n != "" {
		return n
	}
	if n := u.GetUserInfo().GetEmail(); n != "" {
		return n
	}
	if n := u.GetUserInfo().GetPhoneNumber(); n != "" {
		return n
	}
	return ""
}

func UserIdentifier(u *user.User) string {
	if n := u.GetUserInfo().GetUsername(); n != "" {
		return n
	}
	if n := u.GetUserInfo().GetEmail(); n != "" {
		return n
	}
	if n := u.GetUserInfo().GetPhoneNumber(); n != "" {
		return n
	}
	return ""
}
