package protocol

import (
	"errors"
	"fmt"
	"strings"
)

const (
	MinExtendLen = 3   // Extend序列化后最小长度
	MaxExtendLen = 254 // Extend序列化后最大长度
)

// OpenID类型(AuthType)
const (
	AuthTypeNull                = 0x00
	AuthTypeWeb                 = 0x01   // 强登陆skey
	AuthTypeSvr                 = 0x02   // svr访问无需key校验
	AuthTypeLskey               = 0x04   // 弱登陆态lskey
	AuthType3g                  = 0x08   // 老的3G key, 要废掉, 3G的A8key请使用ENUM_AUTH_TYPE_A8
	uthTypeHostkey              = 0x10   // 主人态key
	AuthTypeA8                  = 0x20   // 新的3G A8 key
	AuthTypeA2                  = 0x40   // 新的3G A2 key
	AuthTypeQQOpenID            = 0x100  // 开平QQ登录的openid openkey
	AuthTypeWXOpenID            = 0x200  // 微信登录的openid openkey
	AuthTypeMusicAuthkey        = 0x400  // 音乐authkey
	AuthTypeWXMMkey             = 0x800  // 微信页面mmkey
	AuthTypeIMSdkSelf           = 0x1000 // imsdk k歌自有账号体系
	AuthTypePSkey               = 0x2000 // pskey验证体系(oidb用)
	AuthTypeFacebookOpenID      = 0x2100 // facebook openid
	AuthTypeTwitterOpenID       = 0x2200 // twitter openid
	AuthTypeGMailOpenID         = 0x2400 // twitter openid
	AuthTypeAppleOpenID         = 0x2500 // apple openid
	AuthTypePhoneOpenID         = 0x2510 // 电话号码openid
	AuthTypePhoneOneClickOpenID = 0x2520 // 手机号码一键登录openid
	AuthTypeXtcOpenID           = 0x2530 // 小天才登录的openid openkey
	AuthTypeHuaweiOpenID        = 0x2600 // 华为openid类型
	AuthTypeTclOpenID           = 0x2610 // TCL openid类型
	AuthTypeXiaomiOpenID        = 0x2620 // 小米openid类型
	AuthTypeExtMaxOpenID        = 0x3f00 // end openid
)

// AuthInfo 登录态
type AuthInfo struct {
	UID        uint64 // 用户UID
	TokenType  uint32 // TokenType 前端的登录类型 QMF_PROTOCAL::ENUM_TOKEN_TYPE
	AuthType   uint32 // AuthType tokenType转换之后的authType NS_QZDATA::ENUM_AUTH_TYPE
	OpenID     string // 业务OpenID
	Ticket     string // 业务OpenKey
	ClientIP   uint32 // 小端序列
	AppID      uint32 // 业务类型
	OpenAppID  string // 鉴权相关的AppID
	TraceID    string // 调用链TraceID
	CallerInfo string // 主调服务信息,形如go_ugc_svr
	CalleeInfo string // 被调服务信息,形如l5-xxx-xxx
}

func (a *AuthInfo) Marshal() ([]byte, error) {
	// buf总大小(first section和second section为必填)
	firstSecLen, secondSecLen, thirdSecLen := 0, 0, 0
	if len(a.OpenAppID) != 0 || len(a.OpenID) != 0 || len(a.Ticket) != 0 {
		firstSecLen = len(a.OpenAppID) + len(a.OpenID) + len(a.Ticket) + 4
	}
	if len(a.TraceID) != 0 || len(a.CalleeInfo) != 0 {
		secondSecLen = 1 + len(a.TraceID) + 1 + len(a.CalleeInfo)
	}
	if secondSecLen != 0 && firstSecLen == 0 {
		firstSecLen = 1
	}
	if firstSecLen+secondSecLen > MaxExtendLen {
		return nil, errors.New(fmt.Sprintf("invalid length, first_len:%d, second_len:%d", firstSecLen, secondSecLen))
	}
	buf := make([]byte, MaxExtendLen)

	i := 0
	// marshal section 1
	if firstSecLen > 0 {
		buf[i] = byte(firstSecLen - 1)
		i += 1
		if firstSecLen > 1 {
			buf[i] = byte(len(a.OpenAppID))
			i += 1
			if len(a.OpenAppID) > 0 {
				copy(buf[i:i+len(a.OpenAppID)], a.OpenAppID[:])
				i += len(a.OpenAppID)
			}
			buf[i] = byte(len(a.OpenID))
			i += 1
			if len(a.OpenID) > 0 {
				copy(buf[i:i+len(a.OpenID)], a.OpenID[:])
				i += len(a.OpenID)
			}
			buf[i] = byte(len(a.Ticket))
			i += 1
			if len(a.Ticket) > 0 {
				copy(buf[i:i+len(a.Ticket)], a.Ticket[:])
				i += len(a.Ticket)
			}
		}
	}
	// marshal section 2
	if secondSecLen > 0 {
		buf[i] = byte(secondSecLen - 1)
		i += 1
		if len(a.TraceID) > 0 {
			copy(buf[i:i+len(a.TraceID)], a.TraceID[:])
			i += len(a.TraceID)
		}
		buf[i] = byte('@')
		i += 1
		if len(a.CalleeInfo) > 0 {
			copy(buf[i:i+len(a.CalleeInfo)], a.CalleeInfo[:])
			i += len(a.CalleeInfo)
		}
	}
	// section 1和section 2为必选项,section 3为可选项
	// marshal section 3
	if len(a.CallerInfo) != 0 {
		if firstSecLen == 0 {
			firstSecLen = 1
			buf[i] = byte(firstSecLen - 1)
			i += 1
		}
		if secondSecLen == 0 {
			secondSecLen = 1
			buf[i] = byte(secondSecLen - 1)
			i += 1
		}
		leftLen := MaxExtendLen - firstSecLen - secondSecLen
		if leftLen < 1+len(a.CallerInfo) {
			thirdSecLen = leftLen
		} else {
			thirdSecLen = 1 + len(a.CallerInfo)
		}
		if thirdSecLen > 0 {
			buf[i] = byte(thirdSecLen - 1)
			i += 1
			copy(buf[i:], a.CallerInfo[0:thirdSecLen-1])
		}
	}

	return buf[0 : firstSecLen+secondSecLen+thirdSecLen], nil
}

func (a *AuthInfo) Unmarshal(buf []byte) error {
	if len(buf) == 0 {
		return nil
	}
	// 第一段:OpenAppID, OpenID, Ticket
	i := 0
	firstSecLen := int(buf[i])
	if firstSecLen >= len(buf[i:]) {
		return errors.New(fmt.Sprintf("first exkey len:%d buflen:%d", firstSecLen, len(buf)))
	}
	i += 1
	if firstSecLen > 0 {
		err := a.UnmarshalFirstSection(buf[i : i+firstSecLen])
		if err != nil {
			return err
		}
		i += firstSecLen
	}
	// 第二段:traceid,calleeinfo
	if i < len(buf) {
		secondSecLen := int(buf[i])
		if secondSecLen >= len(buf[i:]) {
			return errors.New(fmt.Sprintf("second exkeylen:%d buflen:%d", secondSecLen, len(buf)))
		}
		i += 1
		if secondSecLen > 0 {
			err := a.UnmarshalSecondSection(buf[i : i+secondSecLen])
			if err != nil {
				return err
			}
			i += secondSecLen
		}
	}
	// 第三段:callerinfo
	if i < len(buf) {
		thirdSecLen := int(buf[i])
		if thirdSecLen >= len(buf[i:]) {
			return errors.New(fmt.Sprintf("third exkeylen:%d buflen:%d", thirdSecLen, len(buf)))
		}
		i += 1
		if thirdSecLen > 0 {
			err := a.UnmarshalThirdSection(buf[i : i+thirdSecLen])
			if err != nil {
				return err
			}
			i += thirdSecLen
		}
	}
	return nil
}

func (a *AuthInfo) UnmarshalFirstSection(buf []byte) error {
	// 总长度
	if len(buf) < MinExtendLen {
		return errors.New(fmt.Sprintf("invalid exkey len:%d", len(buf)))
	}
	i := 0
	// appid长度和内容
	appidLen := int(buf[i])
	i += 1
	if i+appidLen >= len(buf) {
		return errors.New(fmt.Sprintf("invalid appid len:%d, buflen:%d", appidLen, len(buf)))
	}
	if appidLen > 0 {
		a.OpenAppID = string(buf[i : i+appidLen])
		i += appidLen
	}
	// openid长度和内容
	openidLen := int(buf[i])
	i += 1
	if i+openidLen >= len(buf) {
		return errors.New(fmt.Sprintf("invalid openid len:%d, buflen:%d", openidLen, len(buf)))
	}
	if openidLen > 0 {
		a.OpenID = string(buf[i : i+openidLen])
		i += openidLen
	}
	// openkey长度和内容
	openkeyLen := int(buf[i])
	i += 1
	if i+openkeyLen > len(buf) {
		return errors.New(fmt.Sprintf("invalid openkey len:%d, buflen:%d", openkeyLen, len(buf)))
	}
	if openkeyLen > 0 {
		a.Ticket = string(buf[i : i+openkeyLen])
	}
	return nil
}

func (a *AuthInfo) UnmarshalSecondSection(buf []byte) error {
	s := strings.Split(string(buf), "@")
	if len(s) >= 1 {
		a.TraceID = s[0]
	}
	if len(s) >= 2 {
		a.CalleeInfo = s[1]
	}
	return nil
}

func (a *AuthInfo) UnmarshalThirdSection(buf []byte) error {
	a.CallerInfo = string(buf)
	return nil
}
