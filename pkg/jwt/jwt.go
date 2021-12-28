package jwt

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/twinj/uuid"
    "os"
    "time"
    "strconv"
    "net/http"
    "fmt"
    "strings"
    "app/models"
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

type Todo struct {
  UserID uint64 `json:"user_id"`
  Title string `json:"title"`
}

type AccessDetails struct {
    AccessUuid string
    UserId   uint64
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

/**
* Save token to  Redis
*/
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

/**
*  Get token from http header
*/
func ExtractToken(r *http.Request) string {
  bearToken := r.Header.Get("Authorization")
  //normally Authorization the_token_xxx
  strArr := strings.Split(bearToken, " ")
  if len(strArr) == 2 {
     fmt.Println("get header " , strArr[1])
     return strArr[1]
  }
  fmt.Println("can't get header")
  return ""
}

func VerifyToken(r *http.Request) (*jwt.Token, error) {
  tokenString := ExtractToken(r)
  token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
     //Make sure that the token method conform to "SigningMethodHMAC"
     if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
     }
     return []byte(os.Getenv("ACCESS_SECRET")), nil
  })
  if err != nil {
     return nil, err
  }
  return token, nil
}

func TokenValid(r *http.Request) error {
  token, err := VerifyToken(r)
  if err != nil {
     return err
  }
  if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
     return err
  }
  return nil
}

func ExtractTokenMetadata(r *http.Request) (*AccessDetails, error) {
  token, err := VerifyToken(r)
  if err != nil {
    fmt.Println("VerifyToken error")
     return nil, err
  }
  claims, ok := token.Claims.(jwt.MapClaims)
  if ok && token.Valid {
     accessUuid, ok := claims["access_uuid"].(string)
     if !ok {
        fmt.Println("access_uuid error")
        return nil, err
     }
     userId, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["user_id"]), 10, 64)
     if err != nil {
        fmt.Println("user_id error")
        return nil, err
     }
     return &AccessDetails{
        AccessUuid: accessUuid,
        UserId:   userId,
     }, nil
  }
  return nil, err
}

func FetchAuth(authD *AccessDetails) (uint64, error) {
  userid, err := redis.Client.Get(redis.Client.Context(), authD.AccessUuid).Result()
  if err != nil {
     return 0, err
  }
  userID, _ := strconv.ParseUint(userid, 10, 64)
  return userID, nil
}

func DeleteAuth(givenUuid string) (int64,error) {
  deleted, err := redis.Client.Del(redis.Client.Context(), givenUuid).Result()
  if err != nil {
     return 0, err
  }
  return deleted, nil
}

/**
* Refresh Token (non https)
*/
func Refresh_token(refresh_token string) (*TokenDetails){

   var refreshToken models.RefreshTokenObject
   fmt.Println("JWT Refresh_token")
   refreshToken.RefreshToken = refresh_token
   fmt.Println("refresh token  = ", refreshToken.RefreshToken)
   //verify the token
   os.Setenv("REFRESH_SECRET", "mcmvmkmsdnfsdmfdsjf") //this should be in an env file
   token, err := jwt.Parse(refreshToken.RefreshToken, func(token *jwt.Token) (interface{}, error) {
     //Make sure that the token method conform to "SigningMethodHMAC"
     if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
     }
     return []byte(os.Getenv("REFRESH_SECRET")), nil
   })

   //if there is an error, the token must have expired
   if err != nil {
     fmt.Println("JWT Refresh Token error 1")
     return nil
   }

   //is token valid?
   if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
      fmt.Println("JWT Refresh Token error 2")
      return nil
   }

   //Since token is valid, get the uuid:
   claims, ok := token.Claims.(jwt.MapClaims) //the token claims should conform to MapClaims
   if ok && token.Valid {
     refreshUuid, ok := claims["refresh_uuid"].(string) //convert the interface to string
     if !ok {
        return nil
     }
     userId, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["user_id"]), 10, 64)
     if err != nil {
        fmt.Println("JWT Refresh Token error 3")
        return nil
     }
     fmt.Println("refreshUuid ", refreshUuid)
     //Delete the previous Refresh Token
     deleted, delErr := DeleteAuth(refreshUuid)
     if delErr != nil || deleted == 0 { //if any goes wrong
        fmt.Println("JWT Refresh Token error 4")
        return nil
     }

    //Create new pairs of refresh and access tokens
     ts, createErr := CreateToken(userId)
     if  createErr != nil {
       fmt.Println("JWT Refresh Token error 5")
       return nil
     }

     //save the tokens metadata to redis
     saveErr := CreateAuth(userId, ts)
     if saveErr != nil {
        fmt.Println("JWT Refresh Token error 6")
        return nil
     }

     return ts

   } else {
       fmt.Println("JWT Refresh Token error 7")
       return nil
   }
}