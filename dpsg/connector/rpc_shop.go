package connector

import (
	"crypto"
	"crypto/md5"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"golang-project/dpsg/common"
	"golang-project/dpsg/logger"
	"golang-project/dpsg/proto"
	"golang-project/dpsg/rpc"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

func GetCostGem(buynum uint32) uint32 {
	resnum := float32(buynum)
	if 100 >= resnum {
		return 1
	} else if 1000 >= resnum && 100 < resnum {
		return uint32((resnum-100)/((1000-100)/(5.0-1)) + 1)
	} else if 10000 >= resnum && 1000 < resnum {
		return uint32((resnum-1000)/((10000-1000)/(25.0-5)) + 5)
	} else if 100000 >= resnum && 10000 < resnum {
		return uint32((resnum-10000)/((100000-10000)/(125.0-25)) + 25)
	} else if 1000000 >= resnum && 100000 < resnum {
		return uint32((resnum-100000)/((1000000-100000)/(600.0-125)) + 125)
	}

	return uint32((resnum-1000000)/((10000000-1000000)/(3000.0-600)) + 600)
}

func GetWuhunCostGem(buynum uint32) uint32 {
	resnum := float32(buynum)
	if 1 >= resnum {
		return 1
	} else if 10 >= resnum && 5 < resnum {
		return uint32((resnum-1)/((10-1)/(5.0-1)) + 1)
	} else if 100 >= resnum && 25 < resnum {
		return uint32((resnum-10)/((100-10)/(25.0-5)) + 5)
	} else if 1000 >= resnum && 125 < resnum {
		return uint32((resnum-100)/((1000-100)/(125.0-25)) + 25)
	} else if 10000 >= resnum && 600 < resnum {
		return uint32((resnum-1000)/((10000-1000)/(600.0-125)) + 125)
	} else if 100000 >= resnum && 3000 < resnum {
		return uint32((resnum-10000)/((100000-10000)/(3000.0-600)) + 600)
	}

	return uint32((resnum-100000)/((1000000-100000)/(15000-3000)) + 3000)
}

func (self *CNServer) AddMoneyForGM(conn rpc.RpcConn, login rpc.Ping) error {
	ts("CNServer:AddMoneyForGM", conn.GetId())
	defer te("CNServer:AddMoneyForGM", conn.GetId())

	/*self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	p.SetDiamonds(p.GetDiamonds() + 10000000)
	p.SetWuhun(p.GetWuhun() + 10000000)
	p.v.collect_StorageFood(10000000)
	p.v.collect_StorageGold(10000000)*/
	return nil
}

func addDiamond(num uint32, p *player, ResFrom uint32) {
	var tryBuy rpc.TryBuy
	tryBuy.SetType(rpc.ShopItemType_ShopItem_Gem)
	tryBuy.SetNum(num)

	update := p.BuyResource(tryBuy, ResFrom)
	if update != nil {
		WriteResult(p.conn, update)
	}

	//宝石重新同步一次
	//p.SyncPlayerGem()

	notify := &rpc.NotifyRecharge{}
	notify.SetGemNum(num)
	WriteResult(p.conn, notify)
}

func decodeBase64(in string) []byte {
	out := make([]byte, base64.StdEncoding.DecodedLen(len(in)))
	n, err := base64.StdEncoding.Decode(out, []byte(in))
	if err != nil {
		ts("解码Base64出错：", err.Error())
		return nil
	}
	return out[0:n]
}
func (self *CNServer) VerifyGooglePay(conn rpc.RpcConn, gp rpc.GooglePay) (err error) {
	ts("CNServer:VerifyGooglePay", conn.GetId())
	defer te("CNServer:VerifyGooglePay", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return
	}

	ts("Google Pay - 购买数据：", gp.GetData())
	ts("Google Pay - 数据签名：", gp.GetSignature())

	payOrder := common.ReadGooglePayOrder(gp.GetData())
	if p.FindGooglePayNonce(payOrder.DeveloperPayload) == false {
		ts("Google Pay - 非法随机数：", payOrder.DeveloperPayload)
		return
	}

	var googlePay_PublicKey = ""
	var proto_location uint32
	switch gp.GetLocation() {
	case rpc.GameLocation_TaiWan:
		ts("Google Pay - 台湾地区")
		googlePay_PublicKey = common.GetTWGooglePubkey()
		proto_location = proto.Gain_Recharge_TW_GooglePlay
		ts("Google Pay - 公钥：", googlePay_PublicKey)
	}
	if len(googlePay_PublicKey) == 0 {
		ts("Google Pay - 没有公钥")
		return
	}

	ts("Google Pay - 开始验证")
	decodePubkeyBytes := decodeBase64(googlePay_PublicKey)
	if decodePubkeyBytes == nil {
		ts("Google Pay - 错误的公钥：")
		return
	}
	pubkey, err := x509.ParsePKIXPublicKey(decodePubkeyBytes)
	if err != nil {
		ts("Google Pay - 解析公钥出错：", err.Error())
		return
	}

	publicKey := pubkey.(*rsa.PublicKey)
	dataBytes := []byte(gp.GetData())
	signBytes := decodeBase64(gp.GetSignature())

	//dataBytes := []byte(`{"orderId":"12999763169054705758.1363888896840692","packageName":"com.wangcheng.dpsg","productId":"com.wangcheng.dpsg.diamond.500","purchaseTime":1381457150000,"purchaseState":0,"purchaseToken":"scmxtxlucfovimfudrqtheiu.AO-J1OxuWrDH0dGeBUsnyxhs1Xx5x6kmIupdycXsrhfsz8dt7ghTQCNqZqxcQOW_hfsgEw-SEdQESWlmQXMoerELi6kWAE_7Y16S0gp7zxuJBRYl1nlJ83JXd-xk0x5PuUf4GRjw428830Zc3hWv3_0RJ-vxu4kIEQ"}`)
	//signBytes := decodeBase64(`JFOTzJ0UWFBly3uSRtXrZLkBspTZIJT/0r3/gTdvDlgUDgdAALgefHjMbNCu8gz5+mE9tWO7Q4jTc5E1NXKpP/zg5g1pL3Qv1u/hub27oytkmAnxZ845jAjtwRkMz1+AdH/Y/WowuOaLZ7zlI4IAeZ4AsLGyDP+TL9QiXk2cZqpkH4RXLiQjf42kt1w9+60F7UohRLR42L86Cr/ZRorpeOfhS8q9B9swiGNxsq9Ojsejm+RI/fVH4oV2jSDHjNlVfBobku22Dn1O3GvZFrtpvyRTsi5XN3AVytpeO3penUsu5dDtEc9f3ql5ukX8vOmI4PAZnYLpEYlOv2Ytg68gew==`)

	hash := sha1.New()
	hash.Write(dataBytes)
	err1 := rsa.VerifyPKCS1v15(publicKey, crypto.SHA1, hash.Sum(nil), signBytes)
	if err1 != nil {
		ts("Google Pay - 验证失败", err1.Error())
	} else {
		ts("Google Pay - 验证成功")
		p.RemoveGooglePayNonce(payOrder.DeveloperPayload)
		switch payOrder.ProductId {
		case "com.wangcheng.dpsg.gem.500":
			{
				ts("Google Pay - 成功充值500宝石")
				addDiamond(500, p, proto_location)
			}
		case "com.wangcheng.dpsg.gem.1200":
			{
				ts("Google Pay - 成功充值1200宝石")
				addDiamond(1200, p, proto_location)
			}
		case "com.wangcheng.dpsg.gem.2500":
			{
				ts("Google Pay - 成功充值2500宝石")
				addDiamond(2500, p, proto_location)
			}
		case "com.wangcheng.dpsg.gem.6500":
			{
				ts("Google Pay - 成功充值6500宝石")
				addDiamond(6500, p, proto_location)
			}
		case "com.wangcheng.dpsg.gem.14000":
			{
				ts("Google Pay - 成功充值14000宝石")
				addDiamond(14000, p, proto_location)
			}
		}
	}
	return
}
func (self *CNServer) AddGooglePayNonce(conn rpc.RpcConn, gp rpc.GooglePay) (err error) {
	ts("CNServer:AddGooglePayNonce", conn.GetId())
	defer te("CNServer:AddGooglePayNonce", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return
	}

	updatePlayerInfo := &rpc.UpdatePlayerInfo{}
	updatePlayerInfo.SetGooglePayNonce(p.AddGooglePayNonce())
	WriteResult(conn, updatePlayerInfo)

	return
}

func (self *CNServer) Taiwan3Pay(conn rpc.RpcConn, mp rpc.MimigigiPay) (err error) {
	ts("CNServer:Taiwan3Pay", conn.GetId())
	defer te("CNServer:Taiwan3Pay", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return
	}

	var value = ""
	value = common.GetTW3PaySecretKey() + mp.GetUid() + mp.GetTimeStamp() + mp.GetOrderId()

	ts("SecretKey", common.GetTW3PaySecretKey())
	ts("Uid", mp.GetUid())
	ts("TimeStamp", mp.GetTimeStamp())
	ts("OrderId", mp.GetOrderId())
	ts("value", value)

	h := md5.New()
	io.WriteString(h, value)

	hashCode := fmt.Sprintf("%x", h.Sum(nil))

	ts("Token", mp.GetToken())
	ts("hashCode", hashCode)

	if hashCode == mp.GetToken() {
		ts("Taiwan3Pay - 充值验证成功")
		switch mp.GetPayType() {
		case rpc.MimigigiPayType_Tel:
			{
				switch mp.GetPrice() {
				case 150:
					{
						ts("Taiwan3Pay - 电信商充值500宝石")
						addDiamond(500, p, proto.Gain_Recharge_TW_3Pay_Tel)
					}
				case 300:
					{
						ts("Taiwan3Pay - 电信商充值1200宝石")
						addDiamond(1200, p, proto.Gain_Recharge_TW_3Pay_Tel)
					}
				case 590:
					{
						ts("Taiwan3Pay - 电信商充值2500宝石")
						addDiamond(2500, p, proto.Gain_Recharge_TW_3Pay_Tel)
					}
				case 1490:
					{
						ts("Taiwan3Pay - 电信商充值6500宝石")
						addDiamond(6500, p, proto.Gain_Recharge_TW_3Pay_Tel)
					}
				case 2990:
					{
						ts("Taiwan3Pay - 电信商充值14000宝石")
						addDiamond(14000, p, proto.Gain_Recharge_TW_3Pay_Tel)
					}
				}
			}
		case rpc.MimigigiPayType_Gash:
			{
				switch mp.GetPrice() {
				case 50:
					{
						ts("Taiwan3Pay - GASH充值200宝石")
						addDiamond(200, p, proto.Gain_Recharge_TW_3Pay_Gash)
					}
				case 100:
					{
						ts("Taiwan3Pay - GASH充值400宝石")
						addDiamond(400, p, proto.Gain_Recharge_TW_3Pay_Gash)
					}
				case 150:
					{
						ts("Taiwan3Pay - GASH充值600宝石")
						addDiamond(600, p, proto.Gain_Recharge_TW_3Pay_Gash)
					}
				case 300:
					{
						ts("Taiwan3Pay - GASH充值200宝石")
						addDiamond(1200, p, proto.Gain_Recharge_TW_3Pay_Gash)
					}
				case 500:
					{
						ts("Taiwan3Pay - GASH充值2000宝石")
						addDiamond(2000, p, proto.Gain_Recharge_TW_3Pay_Gash)
					}
				case 1000:
					{
						ts("Taiwan3Pay - GASH充值4000宝石")
						addDiamond(4000, p, proto.Gain_Recharge_TW_3Pay_Gash)
					}
				case 3000:
					{
						ts("Taiwan3Pay - GASH充值12000宝石")
						addDiamond(12000, p, proto.Gain_Recharge_TW_3Pay_Gash)
					}
				case 5000:
					{
						ts("Taiwan3Pay - GASH充值20000宝石")
						addDiamond(20000, p, proto.Gain_Recharge_TW_3Pay_Gash)
					}
				}
			}
		}
	} else {
		ts("Taiwan3Pay - 充值验证失败")
	}

	return
}

type IosPayResult struct {
	Status  int
	Receipt IosPayId
}

type IosPayId struct {
	Product_id string
}

func (self *CNServer) VerifyIosPay(conn rpc.RpcConn, pay rpc.IosPay) (err error) {
	ts("CNServer:VerifyIosPay", conn.GetId())
	defer te("CNServer:VerifyIosPay", conn.GetId())

	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return
	}

	rd := make(map[string]string)
	rd["receipt-data"] = pay.GetReceipt()
	b, err := json.Marshal(rd)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	var url string
	if common.IsIosPayTest() {
		url = "https://sandbox.itunes.apple.com/verifyReceipt"
	} else {
		url = "https://buy.itunes.apple.com/verifyReceipt"
	}
	resp, err := http.Post(url, "application/json", strings.NewReader(string(b)))
	if err != nil {
		ts("IosPay call verify failue")
		return
	}

	jsonResult, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	ts("IosPay Result", string(jsonResult))

	var r IosPayResult
	errResult := json.Unmarshal(jsonResult, &r)
	if errResult != nil {
		fmt.Println("errorResult:", errResult)
		return
	}
	if r.Status != 0 {
		ts("IosPay Failue", r.Status)
		return
	}

	ts("IosPay end", r.Receipt.Product_id)

	var proto_location uint32
	switch pay.GetLocation() {
	case rpc.GameLocation_TaiWan:
		{
			proto_location = proto.Gain_Recharge_TW_Ios
		}
	case rpc.GameLocation_English:
		{
			proto_location = proto.Gain_Recharge_EN_Ios
		}
	}
	gemNum := GetGlobalCfg(r.Receipt.Product_id)
	addDiamond(gemNum, p, proto_location)
	ts("Ios Pay - 成功充值宝石")

	return
}

func (self *CNServer) realVerifyIosPayVietnam(conn rpc.RpcConn, pay rpc.IosPayVietnam, bForeTest bool) (err error) {
	ts("CNServer:VerifyIosPayVietnam", conn.GetId())
	defer te("CNServer:VerifyIosPayVietnam", conn.GetId())

	self.l.RLock()
	_, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return
	}

	rd := make(map[string]string)
	rd["receipt-data"] = pay.GetReceipt()
	b, err := json.Marshal(rd)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	var url string
	if common.IsIosPayTest() || bForeTest {
		url = "https://sandbox.itunes.apple.com/verifyReceipt"
	} else {
		url = "https://buy.itunes.apple.com/verifyReceipt"
	}
	resp, err := http.Post(url, "application/json", strings.NewReader(string(b)))
	if err != nil {
		ts("IosPay call verify failue")
		return
	}

	jsonResult, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	ts("IosPay Result", string(jsonResult))

	var r IosPayResult
	errResult := json.Unmarshal(jsonResult, &r)
	if errResult != nil {
		fmt.Println("errorResult:", errResult)
		return
	}

	if r.Status == 21007 && !bForeTest && !common.IsIosPayTest() {
		return self.realVerifyIosPayVietnam(conn, pay, true)
	}

	if r.Status != 0 {
		ts("IosPay Failue", r.Status)
		return
	}

	ts("IosPay end", r.Receipt.Product_id)
	//notify vietnam purchase server about this transaction
	if common.IsIosPayTest() {
		url = "http://test.billing.lienminh.goplay.vn/ConfirmIAP/?"
	} else {
		url = "http://billing.lienminh.goplay.vn/ConfirmIAP/?"
	}
	url += "userid=" + pay.GetUserId() + "&"
	url += "username=" + pay.GetUserName() + "&"
	url += "cpid=" + pay.GetCpid() + "&"
	url += "ipclient=" + conn.GetRemoteIp() + "&"
	url += "storetype=2&"
	url += "productid=" + pay.GetProductId() + "&"
	url += "money=" + pay.GetMoney() + "&"
	url += "storeTransaction=" + pay.GetTransactionId()
	http.Get(url)
	ts("CNServer:Verify ios pay vietnam:", url)
	return
}

func (self *CNServer) VerifyIosPayVietnam(conn rpc.RpcConn, pay rpc.IosPayVietnam) (err error) {
	return self.realVerifyIosPayVietnam(conn, pay, false)
}

func (self *CNServer) BuyResource(conn rpc.RpcConn, buy rpc.TryBuy) (err error) {
	ts("CNServer:BuyResource", conn.GetId())
	defer te("CNServer:BuyResource", conn.GetId())
	if buy.GetType() == rpc.ShopItemType_ShopItem_Gem {
		if common.IsOpenGemBuy() == false {
			return
		}
	}
	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return
	}

	update := p.BuyResource(buy, 0)

	//.need fix 屏蔽购买资源下发消息
	if update != nil {
		WriteResult(conn, update)
	}

	//宝石重新同步一次
	//p.SyncPlayerGem()

	return
}

//来自Center的通知，取回捐赠得兵
func (s *CenterService) NotifyGivePlayerGem(req *proto.NotifyGivePlayerGem, reply *proto.NotifyGivePlayerGemResult) (err error) {
	logger.Info("CenterService.NotifyGivePlayerGem:%s, %d", req.Uid, req.Num)

	cns.l.RLock()
	p, exist := cns.playersbyid[req.Uid]
	cns.l.RUnlock()

	if !exist {
		reply.Ok = false
		return nil
	}

	if p.conn != nil {
		p.conn.Lock()
	}

	defer func() {
		if p.conn != nil {
			p.conn.Unlock()
		}
	}()

	p.GainResource(req.Num, proto.ResType_Gem, proto.Gain_Recharge)
	reply.Ok = true

	// 回到这里的时候，客户端可能断线了，所以要判断下
	if p.conn != nil {
		notify := &rpc.NotifyRecharge{}
		notify.SetGemNum(req.Num)
		WriteResult(p.conn, notify)
	}

	return nil
}

//客户端刷新第三方充值
func (self *CNServer) OnThirdChannelBuyGem(conn rpc.RpcConn, msg rpc.OnThirdChannelBuyGem) error {
	self.l.RLock()
	p, exist := self.players[conn.GetId()]
	self.l.RUnlock()

	if !exist {
		return nil
	}

	p.UpdateThirdGem()

	return nil
}
