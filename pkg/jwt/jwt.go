package jwt

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/twinj/uuid"
    "os"
    "time"
    "strconv"
    redis "app/pkg/redis"
)

type TokenDetails struct {
  AccessToken  string
  RefreshToken string
  AccessUuid   string
  RefreshUuid  string
  AtExpires    int64
  RtExpires    int64
}

/**
*   Create JWT Token
*/
func CreateToken(userid uint64) (*TokenDetails, error) {
  td := &TokenDetails{}
  td.AtExpires = time.Now().Add(time.Minute * 15).Unix()
  td.AccessUuid = uuid.NewV4().String()

  td.RtExpires = time.Now().Add(time.Hour * 24 * 7).Unix()
  td.RefreshUuid = uuid.NewV4().String()

  var err error
  //Creating Access Token
  os.Setenv("ACCESS_SECRET", "jdnfksdmfksd") //this should be in an env file
  atClaims := jwt.MapClaims{}
  atClaims["authorized"] = true
  atClaims["access_uuid"] = td.AccessUuid
  atClaims["user_id"] = userid
  atClaims["exp"] = td.AtExpires
  at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
  td.AccessToken, err = at.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
  if err != nil {
     return nil, err
  }
  //Creating Refresh Token
  os.Setenv("REFRESH_SECRET", "mcmvmkmsdnfsdmfdsjf") //this should be in an env file
  rtClaims := jwt.MapClaims{}
  rtClaims["refresh_uuid"] = td.RefreshUuid
  rtClaims["user_id"] = userid
  rtClaims["exp"] = td.RtExpires
  rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
  td.RefreshToken, err = rt.SignedString([]byte(os.Getenv("REFRESH_SECRET")))
  if err != nil {
     return nil, err
  }
  return td, nil
}

func CreateAuth(userid uint64, td *TokenDetails) error {
    at := time.Unix(td.AtExpires, 0) //converting Unix to UTC(to Time object)
    rt := time.Unix(td.RtExpires, 0)
    now := time.Now()

    errAccess := redis.Client.Set(redis.Client.Context(), td.AccessUuid, strconv.Itoa(int(userid)), at.Sub(now)).Err()
    if errAccess != nil {
        return errAccess
    }
    errRefresh := redis.Client.Set(redis.Client.Context(), td.RefreshUuid, strconv.Itoa(int(userid)), rt.Sub(now)).Err()
    if errRefresh != nil {
        return errRefresh
    }
    return nil
}
