package biligo

import (
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/iyear/biligo/internal/util"
	"github.com/iyear/biligo/proto/dm"
	"github.com/pkg/errors"
	qrcode "github.com/skip2/go-qrcode"
	"github.com/tidwall/gjson"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

type CommClient struct {
	*baseClient
}
type CommSetting struct {

	// 自定义http client
	//
	// 默认为 http.http.DefaultClient
	Client *http.Client

	// Debug模式 true将输出请求信息 false不输出
	//
	// 默认false
	DebugMode bool

	// 自定义UserAgent
	//
	// 默认Chrome随机Agent
	UserAgent string

	// Logger ...
	Logger *log.Logger
}

// NewCommClient
//
// Setting的Auth属性可以随意填写或传入nil，Auth不起到作用，用于访问公共API
func NewCommClient(setting *CommSetting) *CommClient {
	return &CommClient{baseClient: newBaseClient(&baseSetting{
		Client:    setting.Client,
		DebugMode: setting.DebugMode,
		UserAgent: setting.UserAgent,
		Prefix:    "CommClient ",
		Logger:    setting.Logger,
	})}
}

// SetClient
//
// 设置Client,可以用来更换代理等操作
func (c *CommClient) SetClient(client *http.Client) {
	c.client = client
}

// SetUA
//
// 设置UA
func (c *CommClient) SetUA(ua string) {
	c.ua = ua
}

// Raw
//
// base末尾带/
func (c *CommClient) Raw(base, endpoint, method string, payload map[string]string) ([]byte, error) {
	// 不用侵入处理则传入nil
	raw, err := c.raw(base, endpoint, method, payload, nil, nil)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

// RawParse
//
// base末尾带/
func (c *CommClient) RawParse(base, endpoint, method string, payload map[string]string) (*Response, error) {
	raw, err := c.Raw(base, endpoint, method, payload)
	if err != nil {
		return nil, err
	}
	return c.parse(raw)
}

// GetGeoInfo 调用哔哩哔哩API获取地理位置等信息
//
// 会受到自定义 http.Client 代理的影响
func (c *CommClient) GetGeoInfo() (*GeoInfo, error) {
	resp, err := c.RawParse(BiliApiURL,
		"x/web-interface/zone",
		"GET",
		nil,
	)
	if err != nil {
		return nil, err
	}
	var info *GeoInfo
	if err = json.Unmarshal(resp.Data, &info); err != nil {
		return nil, err
	}
	return info, nil
}

// FollowingsGetDetail 获取个人详细的关注列表
//
// pn 页码
//
// ps 每页项数，最大50
func (c *CommClient) FollowingsGetDetail(mid int64, pn int, ps int) (*FollowingsDetail, error) {
	resp, err := c.RawParse(
		BiliApiURL,
		"x/relation/followings",
		"GET",
		map[string]string{
			"vmid": strconv.FormatInt(mid, 10),
			"pn":   strconv.Itoa(pn),
			"ps":   strconv.Itoa(ps),
		},
	)
	if err != nil {
		return nil, err
	}

	var detail = &FollowingsDetail{}
	if err = json.Unmarshal(resp.Data, &detail); err != nil {
		return nil, err
	}
	return detail, nil
}

// VideoGetStat
//
// 获取稿件状态数
func (c *CommClient) VideoGetStat(aid int64) (*VideoSingleStat, error) {
	resp, err := c.RawParse(BiliApiURL,
		"x/web-interface/archive/stat",
		"GET",
		map[string]string{
			"aid": strconv.FormatInt(aid, 10),
		},
	)
	if err != nil {
		return nil, err
	}
	var stat *VideoSingleStat
	if err = json.Unmarshal(resp.Data, &stat); err != nil {
		return nil, err
	}
	return stat, nil
}

// VideoGetInfo 返回视频详细信息，数据较多，可以使用单独的接口获取部分数据
//
// 限制游客访问的视频会返回错误，请使用 BiliClient 发起请求
func (c *CommClient) VideoGetInfo(aid int64) (*VideoInfo, error) {
	resp, err := c.RawParse(BiliApiURL,
		"x/web-interface/view",
		"GET",
		map[string]string{
			"aid": strconv.FormatInt(aid, 10),
		},
	)
	if err != nil {
		return nil, err
	}
	var info *VideoInfo
	if err = json.Unmarshal(resp.Data, &info); err != nil {
		return nil, err
	}
	return info, nil
}

// VideoGetDescription
//
// 获取稿件简介
func (c *CommClient) VideoGetDescription(aid int64) (string, error) {
	resp, err := c.RawParse(BiliApiURL,
		"x/web-interface/archive/desc",
		"GET",
		map[string]string{
			"aid": strconv.FormatInt(aid, 10),
		},
	)
	if err != nil {
		return "", err
	}
	var desc string
	if err = json.Unmarshal(resp.Data, &desc); err != nil {
		return "", err
	}
	return desc, nil
}

// VideoGetPageList
//
// 获取分P列表
func (c *CommClient) VideoGetPageList(aid int64) ([]*VideoPage, error) {
	resp, err := c.RawParse(BiliApiURL,
		"x/player/pagelist",
		"GET",
		map[string]string{
			"aid": strconv.FormatInt(aid, 10),
		},
	)
	if err != nil {
		return nil, err
	}
	var list []*VideoPage
	if err = json.Unmarshal(resp.Data, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// VideoGetOnlineNum
//
// 返回所有终端总计在线观看人数和WEB端在线观看人数 (用类似10万+的文字表示) cid用于分P标识
func (c *CommClient) VideoGetOnlineNum(aid int64, cid int64) (total string, web string, e error) {
	resp, err := c.RawParse(BiliApiURL,
		"x/player/online/total",
		"GET",
		map[string]string{
			"aid": strconv.FormatInt(aid, 10),
			"cid": strconv.FormatInt(cid, 10),
		},
	)
	if err != nil {
		return "", "", err
	}
	var num struct {
		Total string `json:"total,omitempty"`
		Count string `json:"count,omitempty"`
	}
	if err = json.Unmarshal(resp.Data, &num); err != nil {
		return "", "", err
	}
	return num.Total, num.Count, nil
}

// VideoTags
//
// 未登录无法获取 IsAtten,Liked,Hated 字段
func (c *CommClient) VideoTags(aid int64) ([]*VideoTag, error) {
	resp, err := c.RawParse(BiliApiURL,
		"x/tag/archive/tags",
		"GET",
		map[string]string{
			"aid": strconv.FormatInt(aid, 10),
		},
	)
	if err != nil {
		return nil, err
	}
	var tags []*VideoTag
	if err = json.Unmarshal(resp.Data, &tags); err != nil {
		return nil, err
	}
	return tags, nil
}

// VideoGetRecommend 获取视频的相关视频推荐
//
// 最多获取40条推荐视频
func (c *CommClient) VideoGetRecommend(aid int64) ([]*VideoRecommendInfo, error) {
	resp, err := c.RawParse(BiliApiURL,
		"x/web-interface/archive/related",
		"GET",
		map[string]string{
			"aid": strconv.FormatInt(aid, 10),
		},
	)
	if err != nil {
		return nil, err
	}
	var videos []*VideoRecommendInfo
	if err = json.Unmarshal(resp.Data, &videos); err != nil {
		return nil, err
	}
	return videos, nil
}

// VideoGetPlayURL 获取视频取流地址
//
// 所有参数、返回信息和取流方法的说明请直接前往：https://github.com/SocialSisterYi/bilibili-API-collect/blob/master/video/videostream_url.md
func (c *CommClient) VideoGetPlayURL(aid int64, cid int64, qn int, fnval int) (*VideoPlayURLResult, error) {
	resp, err := c.RawParse(
		BiliApiURL,
		"x/player/playurl",
		"GET",
		map[string]string{
			"avid":  strconv.FormatInt(aid, 10),
			"cid":   strconv.FormatInt(cid, 10),
			"qn":    strconv.Itoa(qn),
			"fnval": strconv.Itoa(fnval),
			"fnver": "0",
			"fourk": "1",
		},
	)
	if err != nil {
		return nil, err
	}
	var r *VideoPlayURLResult
	if err = json.Unmarshal(resp.Data, &r); err != nil {
		return nil, err
	}
	return r, nil
}

// VideoShot 获取视频快照
//
// cid属性非必须 传入0表示1P
//
// index为JSON数组截取时间表 true:需要 false:不需要 传入false则Index属性为空
func (c *CommClient) VideoShot(aid int64, cid int64, index bool) (*VideoShot, error) {
	resp, err := c.RawParse(BiliApiURL,
		"x/player/videoshot",
		"GET",
		map[string]string{
			"aid":   strconv.FormatInt(aid, 10),
			"cid":   util.IF(cid == 0, "", strconv.FormatInt(cid, 10)).(string),
			"index": util.IF(index, "1", "0").(string),
		},
	)
	if err != nil {
		return nil, err
	}
	var shot *VideoShot
	if err = json.Unmarshal(resp.Data, &shot); err != nil {
		return nil, err
	}
	return shot, nil
}

// DanmakuGetLikes 获取弹幕点赞数，一次可以获取多条弹幕
//
// Link:https://github.com/SocialSisterYi/bilibili-API-collect/blob/master/danmaku/action.md#%E6%9F%A5%E8%AF%A2%E5%BC%B9%E5%B9%95%E7%82%B9%E8%B5%9E%E6%95%B0
//
// 返回一个map，key为dmid，value为相关信息
// 未登录时UserLike属性恒为0
func (c *CommClient) DanmakuGetLikes(cid int64, dmids []uint64) (map[uint64]*DanmakuGetLikesResult, error) {
	resp, err := c.RawParse(
		BiliApiURL,
		"x/v2/dm/thumbup/stats",
		"GET",
		map[string]string{
			"oid": strconv.FormatInt(cid, 10),
			"ids": util.Uint64SliceToString(dmids, ","),
		},
	)
	if err != nil {
		return nil, err
	}
	var result = make(map[uint64]*DanmakuGetLikesResult)
	for _, dmid := range dmids {
		var r *DanmakuGetLikesResult
		if err = json.Unmarshal([]byte(gjson.Get(string(resp.Data), strconv.FormatUint(dmid, 10)).Raw), &r); err != nil {
			return nil, err
		}
		result[dmid] = r
	}
	return result, nil
}

// GetRelationStat
//
// 获取关系状态数，Whisper和Black恒为0
func (c *CommClient) GetRelationStat(mid int64) (*RelationStat, error) {
	resp, err := c.RawParse(
		BiliApiURL,
		"x/relation/stat",
		"GET",
		map[string]string{
			"vmid": strconv.FormatInt(mid, 10),
		},
	)
	if err != nil {
		return nil, err
	}
	var stat *RelationStat
	if err = json.Unmarshal(resp.Data, &stat); err != nil {
		return nil, err
	}
	return stat, nil
}

// SpaceGetTopArchive
//
// 获取空间置顶稿件
func (c *CommClient) SpaceGetTopArchive(mid int64) (*SpaceVideo, error) {
	resp, err := c.RawParse(
		BiliApiURL,
		"x/space/top/arc",
		"GET",
		map[string]string{
			"vmid": strconv.FormatInt(mid, 10),
		},
	)
	if err != nil {
		return nil, err
	}
	var top *SpaceVideo
	if err = json.Unmarshal(resp.Data, &top); err != nil {
		return nil, err
	}
	return top, nil
}

// SpaceGetMasterpieces
//
// 获取UP代表作
func (c *CommClient) SpaceGetMasterpieces(mid int64) ([]*SpaceVideo, error) {
	resp, err := c.RawParse(
		BiliApiURL,
		"x/space/masterpiece",
		"GET",
		map[string]string{
			"vmid": strconv.FormatInt(mid, 10),
		},
	)
	if err != nil {
		return nil, err
	}
	var mp []*SpaceVideo
	if err = json.Unmarshal(resp.Data, &mp); err != nil {
		return nil, err
	}
	return mp, nil
}

// SpaceGetTags
//
// 获取空间用户个人TAG 上限5条，且内容由用户自定义 带有转义
func (c *CommClient) SpaceGetTags(mid int64) ([]string, error) {
	resp, err := c.RawParse(
		BiliApiURL,
		"x/space/acc/tags",
		"GET",
		map[string]string{
			"mid": strconv.FormatInt(mid, 10),
		},
	)
	if err != nil {
		return nil, err
	}
	// B站这写的是个啥玩意儿
	var tags []string
	for _, tag := range gjson.Get(string(resp.Data), "0.tags").Array() {
		tags = append(tags, tag.String())
	}
	return tags, nil
}

// SpaceGetNotice
//
// 获取空间公告内容
func (c *CommClient) SpaceGetNotice(mid int64) (string, error) {
	resp, err := c.RawParse(
		BiliApiURL,
		"x/space/notice",
		"GET",
		map[string]string{
			"mid": strconv.FormatInt(mid, 10),
		},
	)
	if err != nil {
		return "", err
	}
	// 新建一个变量再 unmarshal 可以把转义部分转回来
	// 直接返回 resp.Data 会带转义符
	var notice string
	if err = json.Unmarshal(resp.Data, &notice); err != nil {
		return "", err
	}
	return notice, nil
}

// SpaceGetLastPlayGame
//
// 获取用户空间近期玩的游戏
func (c *CommClient) SpaceGetLastPlayGame(mid int64) ([]*SpaceGame, error) {
	resp, err := c.RawParse(
		BiliApiURL,
		"x/space/lastplaygame",
		"GET",
		map[string]string{
			"mid": strconv.FormatInt(mid, 10),
		},
	)
	if err != nil {
		return nil, err
	}
	var games []*SpaceGame
	if err = json.Unmarshal(resp.Data, &games); err != nil {
		return nil, err
	}
	return games, nil
}

// SpaceGetLastVideoCoin
//
// 获取用户最近投币的视频明细 如设置隐私查看自己的使用 BiliClient 访问
func (c *CommClient) SpaceGetLastVideoCoin(mid int64) ([]*SpaceVideoCoin, error) {
	resp, err := c.RawParse(
		BiliApiURL,
		"x/space/coin/video",
		"GET",
		map[string]string{
			"vmid": strconv.FormatInt(mid, 10),
		},
	)
	if err != nil {
		return nil, err
	}
	var info []*SpaceVideoCoin
	if err = json.Unmarshal(resp.Data, &info); err != nil {
		return nil, err
	}
	return info, nil
}

// SpaceSearchVideo
//
// 获取用户投稿视频明细
//
// order 排序方式 默认为pubdate 可留空
//
// 最新发布:pubdate
// 最多播放:click
// 最多收藏:stow
//
// tid 筛选分区 0:不进行分区筛选
//
// keyword 关键词 可留空
//
// pn 页码
//
// ps 每页项数
func (c *CommClient) SpaceSearchVideo(mid int64, order string, tid int, keyword string, pn int, ps int) (*SpaceVideoSearchResult, error) {
	resp, err := c.RawParse(
		BiliApiURL,
		"x/space/wbi/arc/search",
		"GET",
		map[string]string{
			"mid":     strconv.FormatInt(mid, 10),
			"order":   order,
			"tid":     strconv.Itoa(tid),
			"keyword": keyword,
			"pn":      strconv.Itoa(pn),
			"ps":      strconv.Itoa(ps),
		},
	)
	if err != nil {
		return nil, err
	}
	var result *SpaceVideoSearchResult
	if err = json.Unmarshal(resp.Data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ChanGet
//
// 获取用户频道列表
func (c *CommClient) ChanGet(mid int64) (*ChannelList, error) {
	resp, err := c.RawParse(
		BiliApiURL,
		"x/space/channel/list",
		"GET",
		map[string]string{
			"mid": strconv.FormatInt(mid, 10),
		},
	)
	if err != nil {
		return nil, err
	}
	var list *ChannelList
	if err = json.Unmarshal(resp.Data, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// ChanGetVideo
//
// 获取用户频道视频
//
// cid 频道ID
//
// pn 页码
//
// ps 每页项数
func (c *CommClient) ChanGetVideo(mid int64, cid int64, pn int, ps int) (*ChanVideo, error) {
	resp, err := c.RawParse(
		BiliApiURL,
		"x/space/channel/video",
		"GET",
		map[string]string{
			"mid": strconv.FormatInt(mid, 10),
			"cid": strconv.FormatInt(cid, 10),
			"pn":  strconv.Itoa(pn),
			"ps":  strconv.Itoa(ps),
		},
	)
	if err != nil {
		return nil, err
	}
	var videos *ChanVideo
	if err = json.Unmarshal(resp.Data, &videos); err != nil {
		return nil, err
	}
	return videos, nil
}

// FavGet
//
// 获取用户的公开收藏夹列表
func (c *CommClient) FavGet(mid int64) (*FavoritesList, error) {
	resp, err := c.RawParse(
		BiliApiURL,
		"x/v3/fav/folder/created/list-all",
		"GET",
		map[string]string{
			"up_mid": strconv.FormatInt(mid, 10),
		},
	)
	if err != nil {
		return nil, err
	}
	var list *FavoritesList
	if err = json.Unmarshal(resp.Data, &list); err != nil {
		return nil, err
	}
	if list == nil {
		return &FavoritesList{}, nil
	}
	return list, nil
}

// FavGetDetail
//
// 获取收藏夹详细信息，部分信息需要登录，请使用 BiliClient 请求
func (c *CommClient) FavGetDetail(mlid int64) (*FavDetail, error) {
	resp, err := c.RawParse(
		BiliApiURL,
		"x/v3/fav/folder/info",
		"GET",
		map[string]string{
			"media_id": strconv.FormatInt(mlid, 10),
		},
	)
	if err != nil {
		return nil, err
	}
	var detail *FavDetail
	if err = json.Unmarshal(resp.Data, &detail); err != nil {
		return nil, err
	}
	return detail, nil
}

// FavGetRes
//
// 获取收藏夹全部内容id 查询权限收藏夹时请使用 BiliClient 请求
func (c *CommClient) FavGetRes(mlid int64) ([]*FavRes, error) {
	resp, err := c.RawParse(
		BiliApiURL,
		"x/v3/fav/resource/ids",
		"GET",
		map[string]string{
			"media_id": strconv.FormatInt(mlid, 10),
		},
	)
	if err != nil {
		return nil, err
	}
	var r []*FavRes
	if err = json.Unmarshal(resp.Data, &r); err != nil {
		return nil, err
	}
	return r, nil
}

// FavGetResDetail 获取收藏夹内容详细内容，带过滤功能
//
// 查询权限收藏夹时请使用 BiliClient 请求
//
// tid 分区id，用于筛选，传入0代表所有分区
//
// keyword 关键词筛选 可留空
//
// order 留空默认按收藏时间
//
// 按收藏时间:mtime
// 按播放量: view
// 按投稿时间：pubtime
//
// tp 内容类型 不知道作用，传入0即可
//
// pn 页码
//
// ps 每页项数 ps不能太大，会报错
func (c *CommClient) FavGetResDetail(mlid int64, tid int, keyword string, order string, tp int, pn int, ps int) (*FavResDetail, error) {
	resp, err := c.RawParse(
		BiliApiURL,
		"x/v3/fav/resource/list",
		"GET",
		map[string]string{
			"media_id": strconv.FormatInt(mlid, 10),
			"tid":      strconv.Itoa(tid),
			"keyword":  keyword,
			"order":    order,
			"type":     strconv.Itoa(tp),
			"ps":       strconv.Itoa(ps),
			"pn":       strconv.Itoa(pn),
		},
	)
	if err != nil {
		return nil, err
	}
	var detail *FavResDetail
	if err = json.Unmarshal(resp.Data, &detail); err != nil {
		return nil, err
	}
	return detail, nil
}

// GetDailyNum
//
// 获取每日分区投稿数
func (c *CommClient) GetDailyNum() (map[int]int, error) {
	resp, err := c.RawParse(BiliApiURL, "x/web-interface/online", "GET", nil)
	if err != nil {
		return nil, err
	}
	var result = make(map[int]int)
	gjson.Get(string(resp.Data), "region_count").ForEach(func(key, value gjson.Result) bool {
		result[int(key.Int())] = int(value.Int())
		return true
	})
	return result, nil
}

// GetUnixNow
//
// 获取服务器的Unix时间戳
func (c *CommClient) GetUnixNow() (int64, error) {
	resp, err := c.RawParse(
		BiliApiURL,
		"x/report/click/now",
		"GET",
		nil,
	)
	if err != nil {
		return -1, err
	}
	var t struct {
		Now int64 `json:"now,omitempty"`
	}
	if err = json.Unmarshal(resp.Data, &t); err != nil {
		return -1, err
	}
	return t.Now, nil
}

// DanmakuGetByPb
//
// 获取实时弹幕(protobuf接口)
func (c *CommClient) DanmakuGetByPb(tp int, cid int64, seg int) (*DanmakuResp, error) {
	resp, err := c.Raw(
		BiliApiURL,
		"x/v2/dm/web/seg.so",
		"GET",
		map[string]string{
			"type":          strconv.Itoa(tp),
			"oid":           strconv.FormatInt(cid, 10),
			"segment_index": strconv.Itoa(seg),
		},
	)
	if err != nil {
		return nil, err
	}
	var reply dm.DmSegMobileReply
	var r = &DanmakuResp{}
	if err := proto.Unmarshal(resp, &reply); err != nil {
		return nil, err
	}
	for _, elem := range reply.GetElems() {
		r.Danmaku = append(r.Danmaku, &Danmaku{
			ID:       uint64(elem.Id),
			Progress: int64(elem.Progress),
			Mode:     int(elem.Mode),
			FontSize: int(elem.Fontsize),
			Color:    int(elem.Color),
			MidHash:  elem.MidHash,
			Content:  elem.Content,
			Ctime:    elem.Ctime,
			Weight:   int(elem.Weight),
			Action:   elem.Action,
			Pool:     int(elem.Pool),
			IDStr:    elem.IdStr,
			Attr:     int(elem.Attr),
		})
	}
	return r, nil

}

// DanmakuGetShot
//
// 获取弹幕快照(最新的几条弹幕)
func (c *CommClient) DanmakuGetShot(aid int64) ([]string, error) {
	resp, err := c.RawParse(
		BiliApiURL,
		"x/v2/dm/ajax",
		"GET",
		map[string]string{
			"aid": strconv.FormatInt(aid, 10),
		},
	)
	if err != nil {
		return nil, err
	}
	var strings []string
	if err = json.Unmarshal(resp.Data, &strings); err != nil {
		return nil, err
	}
	return strings, nil
}

// EmoteGetFreePack 获取免费表情包列表
//
// business 使用场景	reply：评论区 dynamic：动态
//
// 全为免费表情包，如需获取个人专属表情包请使用 BiliClient 请求
func (c *CommClient) EmoteGetFreePack(business string) ([]*EmotePack, error) {
	resp, err := c.RawParse(
		BiliApiURL,
		"x/emote/user/panel/web",
		"GET",
		map[string]string{
			"business": business,
		},
	)
	if err != nil {
		return nil, err
	}
	var pack struct {
		Packages []*EmotePack `json:"packages,omitempty"`
	}
	if err = json.Unmarshal(resp.Data, &pack); err != nil {
		return nil, err
	}
	return pack.Packages, nil
}

// EmoteGetPackDetail 获取指定表情包明细
//
// business 使用场景	reply：评论区 dynamic：动态
//
// ids 多个表情包id的数组
func (c *CommClient) EmoteGetPackDetail(business string, ids []int64) ([]*EmotePack, error) {
	resp, err := c.RawParse(
		BiliApiURL,
		"x/emote/package",
		"GET",
		map[string]string{
			"business": business,
			"ids":      util.Int64SliceToString(ids, ","),
		},
	)
	if err != nil {
		return nil, err
	}
	var packs struct {
		Packages []*EmotePack `json:"packages,omitempty"`
	}
	if err = json.Unmarshal(resp.Data, &packs); err != nil {
		return nil, err
	}
	return packs.Packages, nil
}

// AudioGetInfo
//
// 获取音频信息 部分属性需要登录，请使用 BiliClient 请求
func (c *CommClient) AudioGetInfo(auid int64) (*AudioInfo, error) {
	resp, err := c.RawParse(
		BiliMainURL,
		"audio/music-service-c/web/song/info",
		"GET",
		map[string]string{
			"sid": strconv.FormatInt(auid, 10),
		},
	)
	if err != nil {
		return nil, err
	}
	var info *AudioInfo
	if err = json.Unmarshal(resp.Data, &info); err != nil {
		return nil, err
	}
	return info, nil
}

// AudioGetTags 获取音频TAGs
//
// 根据页面显示观察，应该是歌曲分类
func (c *CommClient) AudioGetTags(auid int64) ([]*AudioTag, error) {
	resp, err := c.RawParse(
		BiliMainURL,
		"audio/music-service-c/web/tag/song",
		"GET",
		map[string]string{
			"sid": strconv.FormatInt(auid, 10),
		},
	)
	if err != nil {
		return nil, err
	}
	var tags []*AudioTag
	if err = json.Unmarshal(resp.Data, &tags); err != nil {
		return nil, err
	}
	return tags, nil
}

// AudioGetMembers
//
// 获取音频创作者信息
func (c *CommClient) AudioGetMembers(auid int64) ([]*AudioMember, error) {
	resp, err := c.RawParse(
		BiliMainURL,
		"audio/music-service-c/web/member/song",
		"GET",
		map[string]string{
			"sid": strconv.FormatInt(auid, 10),
		},
	)
	if err != nil {
		return nil, err
	}
	var members []*AudioMember
	if err = json.Unmarshal(resp.Data, &members); err != nil {
		return nil, err
	}
	return members, nil
}

// AudioGetLyric 获取音频歌词
//
// 同 AudioGetInfo 中的lrc歌词
func (c *CommClient) AudioGetLyric(auid int64) (string, error) {
	resp, err := c.RawParse(
		BiliMainURL,
		"audio/music-service-c/web/song/lyric",
		"GET",
		map[string]string{
			"sid": strconv.FormatInt(auid, 10),
		},
	)
	if err != nil {
		return "", err
	}
	var lrc string
	if err = json.Unmarshal(resp.Data, &lrc); err != nil {
		return "", err
	}
	return lrc, nil
}

// AudioGetStat 获取歌曲状态数
//
// 没有投币数 获取投币数请使用 AudioGetInfo
func (c *CommClient) AudioGetStat(auid int64) (*AudioInfoStat, error) {
	resp, err := c.RawParse(
		BiliMainURL,
		"audio/music-service-c/web/stat/song",
		"GET",
		map[string]string{
			"sid": strconv.FormatInt(auid, 10),
		},
	)
	if err != nil {
		return nil, err
	}
	var stat = &AudioInfoStat{}
	if err = json.Unmarshal(resp.Data, &stat); err != nil {
		return nil, err
	}
	return stat, nil
}

// AudioGetPlayURL 获取音频流URL
//
// 最多获取到
//
// qn 音质
//
// 0 流畅 128K
//
// 1 标准 192K
//
// 2 高品质 320K
//
// 3 无损 FLAC（大会员）
//
// 最高获取到 320K 音质,更高音质请使用 BiliClient 请求
//
// 取流：https://github.com/SocialSisterYi/bilibili-API-collect/blob/master/audio/musicstream_url.md#%E9%9F%B3%E9%A2%91%E6%B5%81%E7%9A%84%E8%8E%B7%E5%8F%96
func (c *CommClient) AudioGetPlayURL(auid int64, qn int) (*AudioPlayURL, error) {
	resp, err := c.RawParse(
		BiliApiURL,
		"audio/music-service-c/url",
		"GET",
		map[string]string{
			"songid":    strconv.FormatInt(auid, 10),
			"quality":   strconv.Itoa(qn),
			"privilege": "2",
			"mid":       "2",
			"platform":  "web",
		},
	)
	if err != nil {
		return nil, err
	}
	var play *AudioPlayURL
	if err = json.Unmarshal(resp.Data, &play); err != nil {
		return nil, err
	}
	return play, nil
}

// ChargeSpaceGetList
//
// 获取用户空间充电名单
func (c *CommClient) ChargeSpaceGetList(mid int64) (*ChargeSpaceList, error) {
	resp, err := c.RawParse(
		BiliElecURL,
		"api/query.rank.do",
		"GET",
		map[string]string{
			"mid": strconv.FormatInt(mid, 10),
		},
	)
	if err != nil {
		return nil, err
	}
	var list *ChargeSpaceList
	if err = json.Unmarshal(resp.Data, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// ChargeVideoGetList
//
// 获取用户视频充电名单
func (c *CommClient) ChargeVideoGetList(mid int64, aid int64) (*ChargeVideoList, error) {
	resp, err := c.RawParse(
		BiliApiURL,
		"x/web-interface/elec/show",
		"GET",
		map[string]string{
			"mid": strconv.FormatInt(mid, 10),
			"aid": strconv.FormatInt(aid, 10),
		},
	)
	if err != nil {
		return nil, err
	}
	var list *ChargeVideoList
	if err = json.Unmarshal(resp.Data, &list); err != nil {
		return nil, err
	}
	return list, nil

}

// LiveGetRoomInfoByMID
//
// 从mid获取直播间信息
func (c *CommClient) LiveGetRoomInfoByMID(mid int64) (*LiveRoomInfoByMID, error) {
	r, err := c.UserGetInfo(mid)
	if err != nil {
		return nil, err
	}
	return (*LiveRoomInfoByMID)(&r.LiveRoom), nil
}

// LiveGetRoomInfoByID 从roomID获取直播间信息
//
// roomID 可为短号也可以是真实房号
func (c *CommClient) LiveGetRoomInfoByID(roomID int64) (*LiveRoomInfoByID, error) {
	resp, err := c.RawParse(
		BiliLiveURL,
		"xlive/web-room/v1/index/getRoomPlayInfo",
		"GET",
		map[string]string{
			"room_id": strconv.FormatInt(roomID, 10),
		},
	)
	if err != nil {
		return nil, err
	}
	var r = &LiveRoomInfoByID{}
	if err = json.Unmarshal(resp.Data, &r); err != nil {
		return nil, err
	}
	return r, nil
}

// LiveGetWsConf 获取直播websocket服务器信息
//
// roomID: 真实直播间ID
func (c *CommClient) LiveGetWsConf(roomID int64) (*LiveWsConf, error) {
	resp, err := c.RawParse(
		BiliLiveURL,
		"room/v1/Danmu/getConf",
		"GET",
		map[string]string{
			"room_id": strconv.FormatInt(roomID, 10),
		},
	)
	if err != nil {
		return nil, err
	}
	var r = &LiveWsConf{}
	if err = json.Unmarshal(resp.Data, &r); err != nil {
		return nil, err
	}
	return r, nil
}

// LiveGetAreaInfo
//
// 获取直播分区信息
func (c *CommClient) LiveGetAreaInfo() ([]*LiveAreaInfo, error) {
	resp, err := c.RawParse(
		BiliLiveURL,
		"room/v1/Area/getList",
		"GET",
		map[string]string{},
	)
	if err != nil {
		return nil, err
	}
	var r []*LiveAreaInfo
	if err = json.Unmarshal(resp.Data, &r); err != nil {
		return nil, err
	}
	return r, nil
}

// LiveGetGuardList 获取直播间大航海列表
//
// roomID: 真实直播间ID
//
// mid: 主播mid
//
// pn: 页码
//
// ps: 每页项数
func (c *CommClient) LiveGetGuardList(roomID int64, mid int64, pn int, ps int) (*LiveGuardList, error) {
	resp, err := c.RawParse(
		BiliLiveURL,
		"xlive/app-room/v1/guardTab/topList",
		"GET",
		map[string]string{
			"roomid":    strconv.FormatInt(roomID, 10),
			"ruid":      strconv.FormatInt(mid, 10),
			"page":      strconv.Itoa(pn),
			"page_size": strconv.Itoa(ps),
		},
	)
	if err != nil {
		return nil, err
	}
	var r = &LiveGuardList{}
	if err = json.Unmarshal(resp.Data, &r); err != nil {
		return nil, err
	}
	return r, nil
}

// LiveGetMedalRank 获取直播间粉丝勋章榜
//
// roomID: 真实直播间ID
//
// mid: 主播mid
func (c *CommClient) LiveGetMedalRank(roomID int64, mid int64) (*LiveMedalRank, error) {
	resp, err := c.RawParse(
		BiliLiveURL,
		"rankdb/v1/RoomRank/webMedalRank",
		"GET",
		map[string]string{
			"roomid": strconv.FormatInt(roomID, 10),
			"ruid":   strconv.FormatInt(mid, 10),
		},
	)
	if err != nil {
		return nil, err
	}
	var r = &LiveMedalRank{}
	if err = json.Unmarshal(resp.Data, &r); err != nil {
		return nil, err
	}
	return r, nil
}

// LiveGetPlayURL 获取直播流信息
//
// qn: 原画:10000 蓝光:400 超清:250 高清:150 流畅:80
func (c *CommClient) LiveGetPlayURL(roomID int64, qn int) (*LivePlayURL, error) {
	resp, err := c.RawParse(
		BiliLiveURL,
		"xlive/web-room/v1/playUrl/playUrl",
		"GET",
		map[string]string{
			"cid":           strconv.FormatInt(roomID, 10),
			"qn":            strconv.Itoa(qn),
			"platform":      "web",
			"https_url_req": "1",
			"ptype":         "16",
		},
	)
	if err != nil {
		return nil, err
	}
	var r = &LivePlayURL{}
	if err = json.Unmarshal(resp.Data, &r); err != nil {
		return nil, err
	}
	return r, nil
}

// LiveGetAllGiftInfo 获取所有礼物信息
//
// areaID: 子分区ID 从 LiveGetAreaInfo 获取
//
// areaParentID: 父分区ID 从 LiveGetAreaInfo 获取
//
// 三个字段可以不用填，但填了有助于减小返回内容的大小，置空(传入0)返回约 2.7w 行，填了三个对应值返回约 1.4w 行
func (c *CommClient) LiveGetAllGiftInfo(roomID int64, areaID int, areaParentID int) (*LiveAllGiftInfo, error) {
	resp, err := c.RawParse(
		BiliLiveURL,
		"xlive/web-room/v1/giftPanel/giftConfig",
		"GET",
		map[string]string{
			"room_id":        strconv.FormatInt(roomID, 10),
			"platform":       "pc",
			"source":         "live",
			"area_id":        strconv.Itoa(areaID),
			"area_parent_id": strconv.Itoa(areaParentID),
		},
	)
	if err != nil {
		return nil, err
	}
	var r = &LiveAllGiftInfo{}
	if err = json.Unmarshal(resp.Data, &r); err != nil {
		return nil, err
	}
	return r, nil
}

// GetEffectConfList ...
func (c *CommClient) GetEffectConfList(roomID int64, areaID int, areaParentID int) (*GetEffectConfList, error) {
	resp, err := c.RawParse(
		BiliLiveURL,
		"xlive/general-interface/v1/fullScSpecialEffect/GetEffectConfList",
		"GET",
		map[string]string{
			"room_id":        strconv.FormatInt(roomID, 10),
			"platform":       "pc",
			"source":         "live",
			"area_id":        strconv.Itoa(areaID),
			"area_parent_id": strconv.Itoa(areaParentID),
		},
	)
	if err != nil {
		return nil, err
	}
	var r = &GetEffectConfList{}
	if err = json.Unmarshal(resp.Data, &r); err != nil {
		return nil, err
	}
	return r, nil
}

// CommentGetCount 获取评论总数
//
// oid: 对应类型的ID
//
// tp: 类型。https://github.com/SocialSisterYi/bilibili-API-collect/tree/master/comment#%E8%AF%84%E8%AE%BA%E5%8C%BA%E7%B1%BB%E5%9E%8B%E4%BB%A3%E7%A0%81
func (c *CommClient) CommentGetCount(oid int64, tp int) (int, error) {
	resp, err := c.RawParse(
		BiliApiURL,
		"x/v2/reply/count",
		"GET",
		map[string]string{
			"oid":  strconv.FormatInt(oid, 10),
			"type": strconv.Itoa(tp),
		},
	)
	if err != nil {
		return -1, err
	}
	var r struct {
		Count int `json:"count"`
	}
	if err = json.Unmarshal(resp.Data, &r); err != nil {
		return -1, err
	}
	return r.Count, nil
}

// CommentGetMain 获取评论区内容
//
// oid: 对应类型的ID
//
// tp: 类型。https://github.com/SocialSisterYi/bilibili-API-collect/tree/master/comment#%E8%AF%84%E8%AE%BA%E5%8C%BA%E7%B1%BB%E5%9E%8B%E4%BB%A3%E7%A0%81
//
// mode: 排序方式
//
// 0 3：仅按热度
//
// 1：按热度+按时间
//
// 2：仅按时间
//
// next: 评论页选择 按热度时：热度顺序页码（0为第一页） 按时间时：时间倒序楼层号
//
// ps: 每页项数
//
// 具体用法请看测试样例
func (c *CommClient) CommentGetMain(oid int64, tp int, mode int, next int, ps int) (*CommentMain, error) {
	resp, err := c.RawParse(
		BiliApiURL,
		"x/v2/reply/main",
		"GET",
		map[string]string{
			"oid":  strconv.FormatInt(oid, 10),
			"type": strconv.Itoa(tp),
			"mode": strconv.Itoa(mode),
			"next": strconv.Itoa(next),
			"ps":   strconv.Itoa(ps),
		},
	)
	if err != nil {
		return nil, err
	}
	var r = &CommentMain{}
	if err = json.Unmarshal(resp.Data, &r); err != nil {
		return nil, err
	}
	return r, nil
}

// CommentGetReply 获取指定评论和二级回复
//
// oid: 对应类型的ID
//
// tp: 类型。https://github.com/SocialSisterYi/bilibili-API-collect/tree/master/comment#%E8%AF%84%E8%AE%BA%E5%8C%BA%E7%B1%BB%E5%9E%8B%E4%BB%A3%E7%A0%81
//
// root: 目标一级评论rpid
//
// pn: 二级评论页码 从1开始
//
// ps: 二级评论每页项数 定义域：1-49
func (c *CommClient) CommentGetReply(oid int64, tp int, root int64, pn int, ps int) (*CommentReply, error) {
	resp, err := c.RawParse(
		BiliApiURL,
		"x/v2/reply/reply",
		"GET",
		map[string]string{
			"oid":  strconv.FormatInt(oid, 10),
			"type": strconv.Itoa(tp),
			"root": strconv.FormatInt(root, 10),
			"pn":   strconv.Itoa(pn),
			"ps":   strconv.Itoa(ps),
		},
	)
	if err != nil {
		return nil, err
	}
	var r = &CommentReply{}
	if err = json.Unmarshal(resp.Data, &r); err != nil {
		return nil, err
	}
	return r, nil
}

func (c *CommClient) UserGetInfo(mid int64) (*UserInfo, error) {
	resp, err := c.RawParse(
		BiliApiURL,
		"x/space/acc/info",
		"GET",
		map[string]string{
			"mid": strconv.FormatInt(mid, 10),
		},
	)
	if err != nil {
		return nil, err
	}
	var r = &UserInfo{}
	if err = json.Unmarshal(resp.Data, &r); err != nil {
		return nil, err
	}
	return r, nil
}

type GetRoomListResp struct {
	NewTags []*TagInfo  `json:"new_tags"`
	List    []*LiveInfo `json:"list"`
	Count   int         `json:"count"`
	HasMore int         `json:"has_more"`
}

// GetRoomList https://api.live.bilibili.com/xlive/web-interface/v1/second/getList?platform=web&parent_area_id=6&area_id=308&sort_type=&page=1&vajra_business_key=
func (c *CommClient) GetRoomList(parentAreaID, areaID, sortType string, page int) (*GetRoomListResp, error) {
	resp, err := c.RawParse(
		BiliLiveURL,
		"xlive/web-interface/v1/second/getList",
		"GET",
		map[string]string{
			"platform":       "web",
			"parent_area_id": parentAreaID,
			"area_id":        areaID,
			"sort_type":      sortType,
			"page":           fmt.Sprint(page),
		},
	)
	if err != nil {
		return nil, err
	}
	var r = &GetRoomListResp{}
	if err = json.Unmarshal(resp.Data, r); err != nil {
		return nil, err
	}
	return r, nil
}

// GetWebAreaList https://api.live.bilibili.com/xlive/web-interface/v1/index/getWebAreaList?source_id=2
func (c *CommClient) GetWebAreaList(sourceID int64) ([]*AreaInfo, error) {
	resp, err := c.RawParse(
		BiliLiveURL,
		"xlive/web-interface/v1/index/getWebAreaList",
		"GET",
		map[string]string{
			"source_id": strconv.FormatInt(sourceID, 10),
		},
	)
	if err != nil {
		return nil, err
	}
	r := struct {
		Data []*AreaInfo `json:"data"`
	}{}
	if err = json.Unmarshal(resp.Data, &r); err != nil {
		return nil, err
	}
	return r.Data, nil
}

// WebQRCodeGenerateResp 二维码链接
type WebQRCodeGenerateResp struct {
	Url       string `json:"url"`
	QrcodeKey string `json:"qrcode_key"`
}

// WebQRCodeGenerate 网页版获取二维码链接
func (c *CommClient) WebQRCodeGenerate() (*WebQRCodeGenerateResp, error) {
	resp, err := c.RawParse(
		BiliPassportURL,
		"x/passport-login/web/qrcode/generate",
		"GET",
		map[string]string{
			"source": "main-fe-header",
		},
	)
	if err != nil {
		return nil, err
	}
	r := &WebQRCodeGenerateResp{}
	if err = json.Unmarshal(resp.Data, r); err != nil {
		return nil, err
	}
	return r, nil
}

// WebQRCodePoolResp 二维码结果
type WebQRCodePoolResp struct {
	Url          string `json:"url"`
	RefreshToken string `json:"refresh_token"`
	Timestamp    int64  `json:"timestamp"` // 毫秒
	Code         int    `json:"code"`
	Message      string `json:"message"`
}

// GetCookieAuth 获取登录信息
func (resp *WebQRCodePoolResp) GetCookieAuth() *CookieAuth {
	auth := &CookieAuth{}
	u, err := url.Parse(resp.Url)
	if err != nil {
		log.Print(err)
		return nil
	}
	for k, v := range u.Query() {
		switch k {
		case "DedeUserID":
			auth.DedeUserID = v[0]
		case "DedeUserID__ckMd5":
			auth.DedeUserIDCkMd5 = v[0]
		case "SESSDATA":
			auth.SESSDATA = v[0]
		case "bili_jct":
			auth.BiliJCT = v[0]
		default:
		}
	}
	return auth
}

// WebQRCodePool 网页版轮询二维码返回结果
func (c *CommClient) WebQRCodePool(qrcodeKey string) (*WebQRCodePoolResp, error) {
	resp, err := c.RawParse(
		BiliPassportURL,
		"x/passport-login/web/qrcode/poll",
		"GET",
		map[string]string{
			"qrcode_key": qrcodeKey,
			"source":     "main-fe-header",
		},
	)
	if err != nil {
		return nil, err
	}
	r := &WebQRCodePoolResp{}
	if err = json.Unmarshal(resp.Data, r); err != nil {
		return nil, err
	}
	return r, nil
}

type QRCodeGetLoginURLResp struct {
	Url      string `json:"url"`
	OauthKey string `json:"oauthKey"`
}

func (resp *QRCodeGetLoginURLResp) ToQRCode() {
	err := qrcode.WriteFile(resp.Url, qrcode.Medium, 256, "qr.png")
	if err != nil {
		panic(err)
	}
}

// QRCodeGetLoginURL 网页版获取二维码链接
func (c *CommClient) QRCodeGetLoginURL() (*QRCodeGetLoginURLResp, error) {
	resp, err := c.RawParse(
		BiliPassportURL,
		"qrcode/getLoginUrl",
		"GET",
		map[string]string{},
	)
	if err != nil {
		return nil, err
	}
	r := &QRCodeGetLoginURLResp{}
	if err = json.Unmarshal(resp.Data, r); err != nil {
		return nil, err
	}
	return r, nil
}

// QRCodeGetLoginInfoResp 返回值
type QRCodeGetLoginInfoResp struct {
	Url          string `json:"url"`
	RefreshToken string `json:"refresh_token"`
	Timestamp    int64  `json:"timestamp"`
}

// GetCookieAuth 获取登录信息
func (resp *QRCodeGetLoginInfoResp) GetCookieAuth() *CookieAuth {
	auth := &CookieAuth{}
	u, err := url.Parse(resp.Url)
	if err != nil {
		log.Print(err)
		return nil
	}
	for k, v := range u.Query() {
		switch k {
		case "DedeUserID":
			auth.DedeUserID = v[0]
		case "DedeUserID__ckMd5":
			auth.DedeUserIDCkMd5 = v[0]
		case "SESSDATA":
			auth.SESSDATA = v[0]
		case "bili_jct":
			auth.BiliJCT = v[0]
		default:
		}
	}
	return auth
}

// QRCodeGetLoginInfo 客户端获取二维码结果
func (c *CommClient) QRCodeGetLoginInfo(oauthKey string) (*QRCodeGetLoginInfoResp, error) {
	resp, err := c.RawParse(
		BiliPassportURL,
		"qrcode/getLoginInfo",
		"POST",
		map[string]string{"oauthKey": oauthKey},
	)
	if err != nil {
		return nil, err
	}

	dataInt, _ := strconv.ParseInt(string(resp.Data), 10, 64)

	switch dataInt {
	case -2:
		return nil, errors.New("Can't Match oauthKey~")
	case -4:
		return nil, errors.New("Can't scan~")
	case -5:
		return nil, errors.New("Can't confirm~")
	default:
	}

	r := &QRCodeGetLoginInfoResp{}
	//fmt.Println(string(resp.Data))
	if err = json.Unmarshal(resp.Data, r); err != nil {
		return nil, err
	}
	return r, nil
}

type GetPopularAnchorRankResp struct {
	List []struct {
		Uid             int64  `json:"uid"`
		Uname           string `json:"uname"`
		Face            string `json:"face"`
		Rank            int    `json:"rank"`
		Score           int    `json:"score"`
		RoomId          int    `json:"room_id"`
		LiveStatus      int    `json:"live_status"`
		Verify          int    `json:"verify"`
		UserNum         int    `json:"user_num"`
		LotStatus       int    `json:"lot_status"`
		RedPocketStatus int    `json:"red_pocket_status"`
		RoomLink        string `json:"room_link"`
	} `json:"list"`
}

func (c *CommClient) GetPopularAnchorRank() (*GetPopularAnchorRankResp, error) {
	resp, err := c.RawParse(
		BiliLiveURL,
		"xlive/general-interface/v1/rank/getPopularAnchorRank",
		"GET",
		map[string]string{},
	)
	if err != nil {
		return nil, err
	}
	r := &GetPopularAnchorRankResp{}
	if err = json.Unmarshal(resp.Data, r); err != nil {
		return nil, err
	}
	return r, nil
}

type GetAreaRankInfoResp struct {
	Items []struct {
		Ruid            int64  `json:"ruid"`
		RoomId          int    `json:"room_id"`
		Uname           string `json:"uname"`
		Face            string `json:"face"`
		Rank            int    `json:"rank"`
		Score           int    `json:"score"`
		Verify          int    `json:"verify"`
		LiveStatus      int    `json:"live_status"`
		LotStatus       int    `json:"lot_status"`
		RedPocketStatus int    `json:"red_pocket_status"`
	} `json:"items"`
	Owner struct {
		Uid          int    `json:"uid"`
		Ruid         int    `json:"ruid"`
		IsFollow     bool   `json:"is_follow"`
		RoomId       int    `json:"room_id"`
		Uname        string `json:"uname"`
		Face         string `json:"face"`
		Rank         int    `json:"rank"`
		Score        int    `json:"score"`
		Verify       int    `json:"verify"`
		LiveStatus   int    `json:"live_status"`
		DiffScore    int    `json:"diff_score"`
		AreaId       int    `json:"area_id"`
		AreaParentId int    `json:"area_parent_id"`
	} `json:"owner"`
	Conf struct {
		Id          int    `json:"id"`
		RankType    int    `json:"rank_type"`
		FeatureType int    `json:"feature_type"`
		RankId      int    `json:"rank_id"`
		RankName    string `json:"rank_name"`
		IconUrl     struct {
			Blue string `json:"blue"`
			Pink string `json:"pink"`
			Grey string `json:"grey"`
		} `json:"icon_url"`
		CycleType      int `json:"cycle_type"`
		ItemDisplayNum int `json:"item_display_num"`
		Rules          []struct {
			Title   string `json:"title"`
			Content string `json:"content"`
		} `json:"rules"`
	} `json:"conf"`
}

func (c *CommClient) GetAreaRankInfo(ruid, confID string) (*GetAreaRankInfoResp, error) {
	resp, err := c.RawParse(
		BiliLiveURL,
		"xlive/general-interface/v1/rank/getAreaRankInfo",
		"GET",
		map[string]string{
			"ruid":    ruid,
			"conf_id": confID,
		},
	)
	if err != nil {
		return nil, err
	}
	r := &GetAreaRankInfoResp{}
	if err = json.Unmarshal(resp.Data, r); err != nil {
		return nil, err
	}
	return r, nil
}

type GetInfoByRoomResp struct {
	RoomInfo struct {
		Uid            int    `json:"uid"`
		RoomId         int    `json:"room_id"`
		ShortId        int    `json:"short_id"`
		Title          string `json:"title"`
		Cover          string `json:"cover"`
		Tags           string `json:"tags"`
		Background     string `json:"background"`
		Description    string `json:"description"`
		LiveStatus     int    `json:"live_status"`
		LiveStartTime  int    `json:"live_start_time"`
		LiveScreenType int    `json:"live_screen_type"`
		LockStatus     int    `json:"lock_status"`
		LockTime       int    `json:"lock_time"`
		HiddenStatus   int    `json:"hidden_status"`
		HiddenTime     int    `json:"hidden_time"`
		AreaId         int    `json:"area_id"`
		AreaName       string `json:"area_name"`
		ParentAreaId   int    `json:"parent_area_id"`
		ParentAreaName string `json:"parent_area_name"`
		Keyframe       string `json:"keyframe"`
		SpecialType    int    `json:"special_type"`
		UpSession      string `json:"up_session"`
		PkStatus       int    `json:"pk_status"`
		IsStudio       bool   `json:"is_studio"`
		Pendants       struct {
			Frame struct {
				Name  string `json:"name"`
				Value string `json:"value"`
				Desc  string `json:"desc"`
			} `json:"frame"`
		} `json:"pendants"`
		OnVoiceJoin int `json:"on_voice_join"`
		Online      int `json:"online"`
		RoomType    struct {
			Field1 int `json:"3-21"`
		} `json:"room_type"`
		SubSessionKey    string      `json:"sub_session_key"`
		LiveId           int         `json:"live_id"`
		LiveIdStr        string      `json:"live_id_str"`
		OfficialRoomId   int         `json:"official_room_id"`
		OfficialRoomInfo interface{} `json:"official_room_info"`
		VoiceBackground  string      `json:"voice_background"`
	} `json:"room_info"`
	AnchorInfo struct {
		BaseInfo struct {
			Uname        string `json:"uname"`
			Face         string `json:"face"`
			Gender       string `json:"gender"`
			OfficialInfo struct {
				Role     int    `json:"role"`
				Title    string `json:"title"`
				Desc     string `json:"desc"`
				IsNft    int    `json:"is_nft"`
				NftDmark string `json:"nft_dmark"`
			} `json:"official_info"`
		} `json:"base_info"`
		LiveInfo struct {
			Level        int    `json:"level"`
			LevelColor   int    `json:"level_color"`
			Score        int    `json:"score"`
			UpgradeScore int    `json:"upgrade_score"`
			Current      []int  `json:"current"`
			Next         []int  `json:"next"`
			Rank         string `json:"rank"`
		} `json:"live_info"`
		RelationInfo struct {
			Attention int `json:"attention"`
		} `json:"relation_info"`
		MedalInfo struct {
			MedalName string `json:"medal_name"`
			MedalId   int    `json:"medal_id"`
			Fansclub  int    `json:"fansclub"`
		} `json:"medal_info"`
		GiftInfo struct {
			Price           int `json:"price"`
			PriceUpdateTime int `json:"price_update_time"`
		} `json:"gift_info"`
	} `json:"anchor_info"`
	NewsInfo struct {
		Uid     int    `json:"uid"`
		Ctime   string `json:"ctime"`
		Content string `json:"content"`
	} `json:"news_info"`
	RankdbInfo struct {
		Roomid    int    `json:"roomid"`
		RankDesc  string `json:"rank_desc"`
		Color     string `json:"color"`
		H5Url     string `json:"h5_url"`
		WebUrl    string `json:"web_url"`
		Timestamp int    `json:"timestamp"`
	} `json:"rankdb_info"`
	AreaRankInfo struct {
		AreaRank struct {
			Index int    `json:"index"`
			Rank  string `json:"rank"`
		} `json:"areaRank"`
		LiveRank struct {
			Rank string `json:"rank"`
		} `json:"liveRank"`
	} `json:"area_rank_info"`
	BattleRankEntryInfo interface{} `json:"battle_rank_entry_info"`
	TabInfo             struct {
		List []struct {
			Type      string `json:"type"`
			Desc      string `json:"desc"`
			IsFirst   int    `json:"isFirst"`
			IsEvent   int    `json:"isEvent"`
			EventType string `json:"eventType"`
			ListType  string `json:"listType"`
			ApiPrefix string `json:"apiPrefix"`
			RankName  string `json:"rank_name"`
		} `json:"list"`
	} `json:"tab_info"`
	ActivityInitInfo struct {
		EventList []interface{} `json:"eventList"`
		WeekInfo  struct {
			BannerInfo interface{} `json:"bannerInfo"`
			GiftName   interface{} `json:"giftName"`
		} `json:"weekInfo"`
		GiftName interface{} `json:"giftName"`
		Lego     struct {
			Timestamp int    `json:"timestamp"`
			Config    string `json:"config"`
		} `json:"lego"`
	} `json:"activity_init_info"`
	VoiceJoinInfo struct {
		Status struct {
			Open        int    `json:"open"`
			AnchorOpen  int    `json:"anchor_open"`
			Status      int    `json:"status"`
			Uid         int    `json:"uid"`
			UserName    string `json:"user_name"`
			HeadPic     string `json:"head_pic"`
			Guard       int    `json:"guard"`
			StartAt     int    `json:"start_at"`
			CurrentTime int    `json:"current_time"`
		} `json:"status"`
		Icons struct {
			IconClose    string `json:"icon_close"`
			IconOpen     string `json:"icon_open"`
			IconWait     string `json:"icon_wait"`
			IconStarting string `json:"icon_starting"`
		} `json:"icons"`
		WebShareLink string `json:"web_share_link"`
	} `json:"voice_join_info"`
	AdBannerInfo struct {
		Data []struct {
			Id                   int         `json:"id"`
			Title                string      `json:"title"`
			Location             string      `json:"location"`
			Position             int         `json:"position"`
			Pic                  string      `json:"pic"`
			Link                 string      `json:"link"`
			Weight               int         `json:"weight"`
			RoomId               int         `json:"room_id"`
			UpId                 int         `json:"up_id"`
			ParentAreaId         int         `json:"parent_area_id"`
			AreaId               int         `json:"area_id"`
			LiveStatus           int         `json:"live_status"`
			AvId                 int         `json:"av_id"`
			IsAd                 bool        `json:"is_ad"`
			AdTransparentContent interface{} `json:"ad_transparent_content"`
			ShowAdIcon           bool        `json:"show_ad_icon"`
		} `json:"data"`
	} `json:"ad_banner_info"`
	SkinInfo struct {
		Id          int    `json:"id"`
		SkinName    string `json:"skin_name"`
		SkinConfig  string `json:"skin_config"`
		ShowText    string `json:"show_text"`
		SkinUrl     string `json:"skin_url"`
		StartTime   int    `json:"start_time"`
		EndTime     int    `json:"end_time"`
		CurrentTime int    `json:"current_time"`
	} `json:"skin_info"`
	WebBannerInfo struct {
		Id               int    `json:"id"`
		Title            string `json:"title"`
		Left             string `json:"left"`
		Right            string `json:"right"`
		JumpUrl          string `json:"jump_url"`
		BgColor          string `json:"bg_color"`
		HoverColor       string `json:"hover_color"`
		TextBgColor      string `json:"text_bg_color"`
		TextHoverColor   string `json:"text_hover_color"`
		LinkText         string `json:"link_text"`
		LinkColor        string `json:"link_color"`
		InputColor       string `json:"input_color"`
		InputTextColor   string `json:"input_text_color"`
		InputHoverColor  string `json:"input_hover_color"`
		InputBorderColor string `json:"input_border_color"`
		InputSearchColor string `json:"input_search_color"`
	} `json:"web_banner_info"`
	LolInfo        interface{} `json:"lol_info"`
	PkInfo         interface{} `json:"pk_info"`
	BattleInfo     interface{} `json:"battle_info"`
	SilentRoomInfo struct {
		Type       string `json:"type"`
		Level      int    `json:"level"`
		Second     int    `json:"second"`
		ExpireTime int    `json:"expire_time"`
	} `json:"silent_room_info"`
	SwitchInfo struct {
		CloseGuard   bool `json:"close_guard"`
		CloseGift    bool `json:"close_gift"`
		CloseOnline  bool `json:"close_online"`
		CloseDanmaku bool `json:"close_danmaku"`
	} `json:"switch_info"`
	RecordSwitchInfo interface{} `json:"record_switch_info"`
	RoomConfigInfo   struct {
		DmText string `json:"dm_text"`
	} `json:"room_config_info"`
	GiftMemoryInfo struct {
		List interface{} `json:"list"`
	} `json:"gift_memory_info"`
	NewSwitchInfo struct {
		RoomSocket           int `json:"room-socket"`
		RoomPropSend         int `json:"room-prop-send"`
		RoomSailing          int `json:"room-sailing"`
		RoomInfoPopularity   int `json:"room-info-popularity"`
		RoomDanmakuEditor    int `json:"room-danmaku-editor"`
		RoomEffect           int `json:"room-effect"`
		RoomFansMedal        int `json:"room-fans_medal"`
		RoomReport           int `json:"room-report"`
		RoomFeedback         int `json:"room-feedback"`
		RoomPlayerWatermark  int `json:"room-player-watermark"`
		RoomRecommendLiveOff int `json:"room-recommend-live_off"`
		RoomActivity         int `json:"room-activity"`
		RoomWebBanner        int `json:"room-web_banner"`
		RoomSilverSeedsBox   int `json:"room-silver_seeds-box"`
		RoomWishingBottle    int `json:"room-wishing_bottle"`
		RoomBoard            int `json:"room-board"`
		RoomSupplication     int `json:"room-supplication"`
		RoomHourRank         int `json:"room-hour_rank"`
		RoomWeekRank         int `json:"room-week_rank"`
		RoomAnchorRank       int `json:"room-anchor_rank"`
		RoomInfoIntegral     int `json:"room-info-integral"`
		RoomSuperChat        int `json:"room-super-chat"`
		RoomTab              int `json:"room-tab"`
		RoomHotRank          int `json:"room-hot-rank"`
		FansMedalProgress    int `json:"fans-medal-progress"`
		GiftBayScreen        int `json:"gift-bay-screen"`
		RoomEnter            int `json:"room-enter"`
		RoomMyIdol           int `json:"room-my-idol"`
		RoomTopic            int `json:"room-topic"`
		FansClub             int `json:"fans-club"`
		RoomPopularRank      int `json:"room-popular-rank"`
		MicUserGift          int `json:"mic_user_gift"`
		NewRoomAreaRank      int `json:"new-room-area-rank"`
		WealthMedal          int `json:"wealth_medal"`
		Bubble               int `json:"bubble"`
		Title                int `json:"title"`
	} `json:"new_switch_info"`
	SuperChatInfo struct {
		Status      int           `json:"status"`
		JumpUrl     string        `json:"jump_url"`
		Icon        string        `json:"icon"`
		RankedMark  int           `json:"ranked_mark"`
		MessageList []interface{} `json:"message_list"`
	} `json:"super_chat_info"`
	OnlineGoldRankInfoV2 struct {
		List []struct {
			Uid         int    `json:"uid"`
			Face        string `json:"face"`
			Uname       string `json:"uname"`
			Score       string `json:"score"`
			Rank        int    `json:"rank"`
			GuardLevel  int    `json:"guard_level"`
			WealthLevel int    `json:"wealth_level"`
		} `json:"list"`
	} `json:"online_gold_rank_info_v2"`
	DmBrushInfo struct {
		MinTime     int `json:"min_time"`
		BrushCount  int `json:"brush_count"`
		SliceCount  int `json:"slice_count"`
		StorageTime int `json:"storage_time"`
	} `json:"dm_brush_info"`
	DmEmoticonInfo struct {
		IsOpenEmoticon   int `json:"is_open_emoticon"`
		IsShieldEmoticon int `json:"is_shield_emoticon"`
	} `json:"dm_emoticon_info"`
	DmTagInfo struct {
		DmTag           int           `json:"dm_tag"`
		Platform        []interface{} `json:"platform"`
		Extra           string        `json:"extra"`
		DmChronosExtra  string        `json:"dm_chronos_extra"`
		DmMode          []interface{} `json:"dm_mode"`
		DmSettingSwitch int           `json:"dm_setting_switch"`
		MaterialConf    interface{}   `json:"material_conf"`
	} `json:"dm_tag_info"`
	TopicInfo struct {
		TopicId   int    `json:"topic_id"`
		TopicName string `json:"topic_name"`
	} `json:"topic_info"`
	GameInfo struct {
		GameStatus int `json:"game_status"`
	} `json:"game_info"`
	WatchedShow struct {
		Switch       bool   `json:"switch"`
		Num          int    `json:"num"`
		TextSmall    string `json:"text_small"`
		TextLarge    string `json:"text_large"`
		Icon         string `json:"icon"`
		IconLocation int    `json:"icon_location"`
		IconWeb      string `json:"icon_web"`
	} `json:"watched_show"`
	TopicRoomInfo struct {
		InteractiveH5Url string `json:"interactive_h5_url"`
		Watermark        int    `json:"watermark"`
	} `json:"topic_room_info"`
	ShowReserveStatus bool `json:"show_reserve_status"`
	SecondCreateInfo  struct {
		ClickPermission  int    `json:"click_permission"`
		CommonPermission int    `json:"common_permission"`
		IconName         string `json:"icon_name"`
		IconUrl          string `json:"icon_url"`
		Url              string `json:"url"`
	} `json:"second_create_info"`
	PlayTogetherInfo struct {
		Switch   int           `json:"switch"`
		IconList []interface{} `json:"icon_list"`
	} `json:"play_together_info"`
	CloudGameInfo struct {
		IsGaming int `json:"is_gaming"`
	} `json:"cloud_game_info"`
	LikeInfoV3 struct {
		TotalLikes    int      `json:"total_likes"`
		ClickBlock    bool     `json:"click_block"`
		CountBlock    bool     `json:"count_block"`
		GuildEmoText  string   `json:"guild_emo_text"`
		GuildDmText   string   `json:"guild_dm_text"`
		LikeDmText    string   `json:"like_dm_text"`
		HandIcons     []string `json:"hand_icons"`
		DmIcons       []string `json:"dm_icons"`
		EggshellsIcon string   `json:"eggshells_icon"`
		CountShowTime int      `json:"count_show_time"`
		ProcessIcon   string   `json:"process_icon"`
		ProcessColor  string   `json:"process_color"`
	} `json:"like_info_v3"`
	LivePlayInfo struct {
		ShowWidgetBanner bool `json:"show_widget_banner"`
		ShowLeftEntry    bool `json:"show_left_entry"`
	} `json:"live_play_info"`
	MultiVoice struct {
		SwitchStatus int           `json:"switch_status"`
		Members      []interface{} `json:"members"`
		MvRole       int           `json:"mv_role"`
		SeatType     int           `json:"seat_type"`
		InvokingTime int           `json:"invoking_time"`
		Version      int           `json:"version"`
		Pk           interface{}   `json:"pk"`
		BizSessionId string        `json:"biz_session_id"`
	} `json:"multi_voice"`
	PopularRankInfo struct {
		Rank       int    `json:"rank"`
		Countdown  int    `json:"countdown"`
		Timestamp  int    `json:"timestamp"`
		Url        string `json:"url"`
		OnRankName string `json:"on_rank_name"`
		RankName   string `json:"rank_name"`
	} `json:"popular_rank_info"`
	NewAreaRankInfo struct {
		Items []struct {
			ConfId      int    `json:"conf_id"`
			RankName    string `json:"rank_name"`
			Uid         int    `json:"uid"`
			Rank        int    `json:"rank"`
			IconUrlBlue string `json:"icon_url_blue"`
			IconUrlPink string `json:"icon_url_pink"`
			IconUrlGrey string `json:"icon_url_grey"`
			JumpUrlLink string `json:"jump_url_link"`
			JumpUrlPc   string `json:"jump_url_pc"`
			JumpUrlPink string `json:"jump_url_pink"`
			JumpUrlWeb  string `json:"jump_url_web"`
		} `json:"items"`
		RotationCycleTimeWeb int `json:"rotation_cycle_time_web"`
	} `json:"new_area_rank_info"`
	GiftStar struct {
		Show                 bool `json:"show"`
		DisplayWidgetAbGroup int  `json:"display_widget_ab_group"`
	} `json:"gift_star"`
	ProgressForWidget struct {
		GiftStarProcess struct {
			TaskInfo struct {
				StartDate   int `json:"start_date"`
				ProcessList []struct {
					GiftId       int    `json:"gift_id"`
					GiftImg      string `json:"gift_img"`
					GiftName     string `json:"gift_name"`
					CompletedNum int    `json:"completed_num"`
					TargetNum    int    `json:"target_num"`
				} `json:"process_list"`
				Finished       bool   `json:"finished"`
				DdlTimestamp   int    `json:"ddl_timestamp"`
				Version        int64  `json:"version"`
				RewardGift     int    `json:"reward_gift"`
				RewardGiftImg  string `json:"reward_gift_img"`
				RewardGiftName string `json:"reward_gift_name"`
			} `json:"task_info"`
			PreloadTimestamp int         `json:"preload_timestamp"`
			Preload          bool        `json:"preload"`
			PreloadTaskInfo  interface{} `json:"preload_task_info"`
			WidgetBg         string      `json:"widget_bg"`
			JumpSchema       string      `json:"jump_schema"`
			AbGroup          int         `json:"ab_group"`
		} `json:"gift_star_process"`
		WishProcess interface{} `json:"wish_process"`
	} `json:"progress_for_widget"`
	RevenueDemotion struct {
		GlobalGiftConfigDemotion bool `json:"global_gift_config_demotion"`
	} `json:"revenue_demotion"`
	RevenueMaterialMd5  interface{} `json:"revenue_material_md5"`
	VideoConnectionInfo interface{} `json:"video_connection_info"`
	PlayerThrottleInfo  struct {
		Status              int `json:"status"`
		NormalSleepTime     int `json:"normal_sleep_time"`
		FullscreenSleepTime int `json:"fullscreen_sleep_time"`
		TabSleepTime        int `json:"tab_sleep_time"`
		PromptTime          int `json:"prompt_time"`
	} `json:"player_throttle_info"`
	GuardInfo struct {
		Count                   int `json:"count"`
		AnchorGuardAchieveLevel int `json:"anchor_guard_achieve_level"`
	} `json:"guard_info"`
	HotRankInfo interface{} `json:"hot_rank_info"`
}

func (c *CommClient) GetInfoByRoom(roomID int64) (*GetInfoByRoomResp, error) {
	resp, err := c.RawParse(
		BiliLiveURL,
		"xlive/web-room/v1/index/getInfoByRoom",
		"GET",
		map[string]string{
			"room_id": fmt.Sprint(roomID),
		},
	)
	if err != nil {
		return nil, err
	}
	r := &GetInfoByRoomResp{}
	if err = json.Unmarshal(resp.Data, r); err != nil {
		return nil, err
	}
	return r, nil
}

type GetOnlineGoldRankResp struct {
	OnlineNum      int `json:"onlineNum"`
	OnlineRankItem []struct {
		UserRank  int    `json:"userRank"`
		Uid       int    `json:"uid"`
		Name      string `json:"name"`
		Face      string `json:"face"`
		Score     int    `json:"score"`
		MedalInfo *struct {
			GuardLevel       int    `json:"guardLevel"`
			MedalColorStart  int    `json:"medalColorStart"`
			MedalColorEnd    int    `json:"medalColorEnd"`
			MedalColorBorder int    `json:"medalColorBorder"`
			MedalName        string `json:"medalName"`
			Level            int    `json:"level"`
			TargetId         int64  `json:"targetId"`
			IsLight          int    `json:"isLight"`
		} `json:"medalInfo"`
		GuardLevel  int `json:"guard_level"`
		WealthLevel int `json:"wealth_level"`
	} `json:"OnlineRankItem"`
	OwnInfo struct {
		Uid         int    `json:"uid"`
		Name        string `json:"name"`
		Face        string `json:"face"`
		Rank        int    `json:"rank"`
		NeedScore   int    `json:"needScore"`
		Score       int    `json:"score"`
		GuardLevel  int    `json:"guard_level"`
		WealthLevel int    `json:"wealth_level"`
	} `json:"ownInfo"`
	TipsText  string `json:"tips_text"`
	ValueText string `json:"value_text"`
	Ab        struct {
		GuardAccompanyList int `json:"guard_accompany_list"`
	} `json:"ab"`
}

func (c *CommClient) GetOnlineGoldRank(rUID, roomID, page, pageSize int64) (*GetOnlineGoldRankResp, error) {
	resp, err := c.RawParse(
		BiliLiveURL,
		"xlive/general-interface/v1/rank/getOnlineGoldRank",
		"GET",
		map[string]string{
			"ruid":     fmt.Sprint(rUID),
			"roomId":   fmt.Sprint(roomID),
			"page":     fmt.Sprint(page),
			"pageSize": fmt.Sprint(pageSize),
		},
	)
	if err != nil {
		return nil, err
	}
	r := &GetOnlineGoldRankResp{}
	if err = json.Unmarshal(resp.Data, r); err != nil {
		return nil, err
	}
	return r, nil
}

type GetUserExResp struct {
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Message string `json:"message"`
	Data    struct {
		User struct {
			Role            int `json:"role"`
			UserLevel       int `json:"user_level"`
			MasterLevel     int `json:"master_level"`
			NextMasterLevel int `json:"next_master_level"`
			NeedMasterScore int `json:"need_master_score"`
			MasterRank      int `json:"master_rank"`
			Verify          int `json:"verify"`
		} `json:"user"`
		Feed struct {
			FansCount   int `json:"fans_count"`
			FeedCount   int `json:"feed_count"`
			IsFollowed  int `json:"is_followed"`
			IsFollowing int `json:"is_following"`
		} `json:"feed"`
		Room struct {
			LiveStatus  int    `json:"live_status"`
			RoomId      int    `json:"room_id"`
			ShortRoomId int    `json:"short_room_id"`
			Title       string `json:"title"`
			Cover       string `json:"cover"`
			Keyframe    string `json:"keyframe"`
			Online      int    `json:"online"`
			RoomLink    string `json:"room_link"`
		} `json:"room"`
		Uid string `json:"uid"`
	} `json:"data"`
}

func (c *CommClient) GetUserEx(uid int64) (*GetUserExResp, error) {
	resp, err := http.Get("https://api.vc.bilibili.com/user_ex/v1/user/detail?uid=" + fmt.Sprint(uid) + "&user[]=role&user[]=level&room[]=live_status&room[]=room_link&feed[]=fans_count&feed[]=feed_count&feed[]=is_followed&feed[]=is_following&platform=pc")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	r := &GetUserExResp{}
	if err = json.Unmarshal(bytes, r); err != nil {
		return nil, err
	}
	return r, nil
}

type QueryAppDetailRsp struct {
	AppInfo struct {
		AppId                int64  `json:"app_id"`
		Name                 string `json:"name"`
		Icon                 string `json:"icon"`
		PromotionCover       string `json:"promotion_cover"`
		Type                 int    `json:"type"`
		Status               int    `json:"status"`
		Version              string `json:"version"`
		DeveloperId          int    `json:"developer_id"`
		DeveloperName        string `json:"developer_name"`
		Abstract             string `json:"abstract"`
		DetailedIntroduction string `json:"detailed_introduction"`
		IsDefault            bool   `json:"is_default"`
		Own                  bool   `json:"own"`
		IconMark             string `json:"icon_mark"`
		DownloadUrl          string `json:"download_url"`
		LikeCount            int    `json:"like_count"`
		IsUsingAnchors       []struct {
			Avatar string `json:"avatar"`
			Name   string `json:"name"`
			RoomId int    `json:"room_id"`
			Cover  string `json:"cover"`
			Title  string `json:"title"`
		} `json:"is_using_anchors"`
		Level              int         `json:"level"`
		PublishTime        int         `json:"publish_time"`
		DeveloperFace      string      `json:"developer_face"`
		RecommendImage     string      `json:"recommend_image"`
		IsLike             bool        `json:"is_like"`
		Id                 int         `json:"id"`
		PcliveCover        string      `json:"pclive_cover"`
		IsHq               int         `json:"is_hq"`
		OwnCount           int         `json:"own_count"`
		TagList            interface{} `json:"tag_list"`
		ArchiveBvid        string      `json:"archive_bvid"`
		UseInstruction     string      `json:"use_instruction"`
		PreviewCovers      []string    `json:"preview_covers"`
		Description        string      `json:"description"`
		IsSupportObs       bool        `json:"is_support_obs"`
		DeveloperBName     string      `json:"developer_b_name"`
		DeveloperSign      string      `json:"developer_sign"`
		DeveloperBLevel    int         `json:"developer_b_level"`
		DeveloperLikeCount int         `json:"developer_like_count"`
		DeveloperOwnCount  int         `json:"developer_own_count"`
		CreateTime         int         `json:"create_time"`
	} `json:"app_info"`
}

func (c *CommClient) QueryAppDetail(app_id int64) (*QueryAppDetailRsp, error) {
	resp, err := c.RawParse(
		BiliLiveURL,
		"xlive/virtual-interface/v2/app/queryAppDetail",
		"GET",
		map[string]string{
			"app_id": fmt.Sprint(app_id),
		},
	)
	if err != nil {
		return nil, err
	}
	r := &QueryAppDetailRsp{}
	if err = json.Unmarshal(resp.Data, r); err != nil {
		return nil, err
	}
	return r, nil
}
