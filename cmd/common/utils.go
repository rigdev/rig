package common

import (
	"context"
	"fmt"
	"math"
	"net/mail"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/distribution/reference"
	"github.com/rigdev/rig-go-api/api/v1/group"
	"github.com/rigdev/rig-go-api/api/v1/service_account"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/robfig/cron"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/validation"
)

func ValidateAll(_ string) error {
	return nil
}

func BoolValidate(bool string) error {
	if bool != "true" && bool != "false" {
		return errors.InvalidArgumentErrorf("invalid boolean value")
	}
	return nil
}

func ValidateInt(input string) error {
	_, err := strconv.Atoi(input)
	if err != nil {
		return err
	}
	return nil
}

func ValidateUInt(input string) error {
	n, err := strconv.Atoi(input)
	if err != nil {
		return err
	}

	if n < 0 {
		return errors.New("cannot be negative")
	}

	return nil
}

func ValidateIntInRange(minInclusive, maxInclusive int) func(string) error {
	return func(s string) error {
		n, err := strconv.Atoi(s)
		if err != nil {
			return err
		}
		if !(minInclusive <= n && n <= maxInclusive) {
			return fmt.Errorf("must be between %v and %v", minInclusive, maxInclusive)
		}
		return nil
	}
}

func ValidateNonEmpty(input string) error {
	if input == "" {
		return errors.InvalidArgumentErrorf("value cannot be empty")
	}
	return nil
}

func ValidateEmail(input string) error {
	_, err := mail.ParseAddress(input)
	if err != nil {
		return err
	}
	return nil
}

func ValidateSystemName(input string) error {
	if l := len(input); l < 3 || l > 63 {
		return errors.InvalidArgumentErrorf("must be between 3 and 63 characters long")
	}

	if !regexp.MustCompile(`^[a-z][a-z0-9-]+[a-z0-9]$`).MatchString(input) {
		return errors.InvalidArgumentErrorf("invalid name; can only contain a-z, 0-9 and '-'")
	}

	return nil
}

func ValidateKubernetesName(input string) error {
	if errs := validation.IsQualifiedName(input); errs != nil {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

func ValidateURL(input string) error {
	_, err := url.Parse(input)
	return err
}

func ValidateAbsolutePath(input string) error {
	if abs := path.IsAbs(input); !abs {
		return errors.InvalidArgumentErrorf("must be an absolute path")
	}
	return nil
}

func ValidateFilePath(input string) error {
	if path.Ext(input) == "" {
		return errors.InvalidArgumentErrorf("must be a file path")
	}
	return nil
}

func ValidateImage(input string) error {
	_, err := reference.ParseDockerRef(input)
	if err != nil {
		return err
	}

	return nil
}

func ValidateBool(s string) error {
	if s == "" {
		return nil
	}

	if _, err := parseBool(s); err != nil {
		return err
	}

	return nil
}

func ValidateQuantity(s string) error {
	_, err := resource.ParseQuantity(s)
	return err
}

func ValidatePort(s string) error {
	n, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	if n <= 0 || n >= 65536 {
		return errors.New("port number x must be 0 < x < 65536")
	}
	return nil
}

func ValidateUnique(values []string) func(string) error {
	return func(s string) error {
		for _, v := range values {
			if v == s {
				return errors.New("must be unique")
			}
		}
		return nil
	}
}

func ValidateLength(minLength, maxLength int) func(string) error {
	return func(s string) error {
		if !(minLength <= len(s) && len(s) <= maxLength) {
			return fmt.Errorf("length must be betwen %v and %v", minLength, maxLength)
		}
		return nil
	}
}

func ValidateCronExpression(s string) error {
	_, err := cron.Parse(s)
	return err
}

func ValidateAnd(validators ...func(s string) error) func(s string) error {
	return func(s string) error {
		for _, v := range validators {
			if err := v(s); err != nil {
				return err
			}
		}
		return nil
	}
}

func ValidateAllowEmpty(validator func(string) error) func(s string) error {
	return func(s string) error {
		if len(s) == 0 {
			return nil
		}
		return validator(s)
	}
}

func ValidateDuration(s string) error {
	_, err := time.ParseDuration(s)
	return err
}

func parseBool(s string) (bool, error) {
	switch s {
	case "1", "t", "T", "true", "TRUE", "True", "y", "Y", "yes", "YES", "Yes":
		return true, nil
	case "0", "f", "F", "false", "FALSE", "False", "n", "N", "no", "NO", "No":
		return false, nil
	}
	return false, errors.InvalidArgumentErrorf("invalid bool format")
}

func GetMember(ctx context.Context, rc rig.Client, p Prompter) (string, string, []string, error) {
	i, _, err := p.Select("Select Member Type", []string{"User", "Service Account"})
	if err != nil {
		return "", "", nil, err
	}
	if i == 0 {
		u, userID, err := GetUser(ctx, "", rc, p)
		if err != nil {
			return "", "", nil, err
		}
		id, err := uuid.Parse(userID)
		if err != nil {
			return "", "", nil, err
		}
		return string(id), "", u.UserInfo.GetGroupIds(), nil
	} else if i == 1 {
		saResp, err := rc.ServiceAccount().List(ctx, &connect.Request[service_account.ListRequest]{
			Msg: &service_account.ListRequest{},
		})
		if err != nil {
			return "", "", nil, err
		}

		var sas []string
		for _, sa := range saResp.Msg.GetServiceAccounts() {
			sas = append(sas, sa.GetName())
		}

		i, _, err := p.Select("Select Service Account", sas)
		if err != nil {
			return "", "", nil, err
		}

		serviceAccount := saResp.Msg.GetServiceAccounts()[i]
		id, err := uuid.Parse(serviceAccount.GetServiceAccountId())
		if err != nil {
			return "", "", nil, err
		}
		return "", string(id), serviceAccount.GetGroupIds(), nil
	}

	return "", "", nil, nil
}

func GetUser(ctx context.Context, identifier string, rc rig.Client, p Prompter) (*user.User, string, error) {
	var err error
	if identifier == "" {
		identifier, err = p.Input("User Identifier:", ValidateNonEmptyOpt)
		if err != nil {
			return nil, "", err
		}
	}
	var u *user.User
	var resID string
	id, err := uuid.Parse(identifier)
	if err != nil {
		ident, err := ParseUserIdentifier(identifier)
		if err != nil {
			return nil, "", err
		}

		res, err := rc.User().GetByIdentifier(ctx, connect.NewRequest(&user.GetByIdentifierRequest{
			Identifier: ident,
		}))
		if err != nil {
			return nil, "", err
		}
		resID = res.Msg.GetUser().GetUserId()
		u = res.Msg.GetUser()
	} else {
		res, err := rc.User().Get(ctx, connect.NewRequest(&user.GetRequest{
			UserId: id.String(),
		}))
		if err != nil {
			return nil, "", err
		}

		u = res.Msg.GetUser()
		resID = id.String()
	}
	return u, resID, nil
}

func GetGroup(ctx context.Context, id string, nc rig.Client, p Prompter) (*group.Group, string, error) {
	var err error
	if id == "" {
		id, err = p.Input("Group Id:", ValidateSystemNameOpt)
		if err != nil {
			return nil, "", err
		}
	}
	var g *group.Group
	res, err := nc.Group().Get(ctx, connect.NewRequest(&group.GetRequest{
		GroupId: id,
	}))
	if err != nil {
		return nil, "", err
	}
	g = res.Msg.GetGroup()
	return g, id, nil
}

func FormatField(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, " ", "-"))
}

func FormatIntToSI(n uint64, decimals int) string {
	scale := uint64(math.Pow10(decimals))
	n = (n * scale) / scale

	suffix := ""
	if n < 1_000 {
		scale = 1
		suffix = ""
	} else if n < 1_000_000 {
		scale = 1_000
		suffix = "k"
	} else if n < 1_000_000_000 {
		scale = 1_000_000
		suffix = "M"
	} else if n < 1_000_000_000_000 {
		scale = 1_000_000_000
		suffix = "G"
	} else {
		scale = 1_000_000_000_000
		suffix = "T"
	}

	result := float64(n) / float64(scale)
	return ToStringWithSignificantDigits(result, 3) + suffix
}

func ToStringWithSignificantDigits(f float64, digits int) string {
	sign := ""
	if f < 0 {
		sign = "-"
	}

	ff := math.Abs(f)
	exponent := int(math.Max(0, math.Ceil(math.Log10(ff))))
	ff = math.Round((ff / math.Pow10(exponent)) * math.Pow10(digits))

	s := strconv.FormatInt(int64(ff), 10)
	if s == "0" {
		return "0"
	}

	if len(s) < exponent {
		return sign + s + strings.Repeat("0", exponent-len(s))
	}

	integerPart := s[:exponent]
	if exponent == 0 {
		integerPart = "0"
	}

	fractionalPart := s[exponent:]
	if len(s) < digits {
		fractionalPart = strings.Repeat("0", digits-len(s)) + fractionalPart
	}

	fractionIsOnlyZeros := strings.Count(fractionalPart, "0") == len(fractionalPart)
	if fractionIsOnlyZeros {
		fractionalPart = ""
	} else {
		fractionalPart = "." + fractionalPart
	}

	return sign + integerPart + fractionalPart
}

func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprint(d.Truncate(time.Second))
	}
	if d < time.Hour {
		minutes := int(math.Floor(d.Minutes()))
		seconds := int(math.Floor(d.Seconds())) % 60
		return fmt.Sprintf("%vm %vs", minutes, seconds)
	}
	if d < 48*time.Hour { // Just to have a bit more precision between 24 and 48 hours
		hours := int(math.Floor(d.Hours()))
		minutes := int(math.Floor(d.Minutes())) % 60
		return fmt.Sprintf("%vh %vm", hours, minutes)
	}
	days := int(math.Floor(d.Hours())) / 24
	hours := int(math.Floor(d.Hours())) % 24
	return fmt.Sprintf("%vd %vh", days, hours)
}

func GetAllAnnotations(cmd *cobra.Command) map[string]string {
	res := make(map[string]string)
	for p := cmd; p != nil; p = p.Parent() {
		for k, v := range p.Annotations {
			if _, ok := res[k]; ok {
				continue
			}
			res[k] = v
		}
	}
	return res
}

func FormatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.DateTime)
}
