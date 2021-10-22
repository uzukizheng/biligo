package biligo

import (
	"testing"
)

var testCommClient = newTestCommClient()

func newTestCommClient() *CommClient {
	c := NewCommClient(&CommSetting{
		DebugMode: true,
	})

	return c
}
func TestCommClient_GetGeoInfo(t *testing.T) {
	info, err := testCommClient.GetGeoInfo()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Logf("ip: %s,country: %s,province: %s,city: %s,isp: %s,latitude: %.2f,longitude: %.2f,zoneID: %d,countryCode: %d",
		info.Addr, info.Country, info.Province, info.City, info.Isp, info.Latitude, info.Longitude, info.ZoneID, info.CountryCode)
}
func TestCommClient_VideoStatus(t *testing.T) {
	stat, err := testCommClient.VideoGetStat(759949922)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Logf("aid: %d,bvid: %s,view: %d,danmaku: %d,reply: %d,favorite: %d,coin: %d,share: %d,now_rank: %d,his_rank: %d,like: %d,dislike: %d,no_reprint: %d,copyright: %d,argue_msg: %s,evaluation: %s",
		stat.AID, stat.BVID, stat.View, stat.Danmaku, stat.Reply, stat.Favorite, stat.Coin, stat.Share, stat.NowRank, stat.HisRank, stat.Like, stat.Dislike, stat.NoReprint, stat.Copyright, stat.ArgueMsg, stat.Evaluation)
}
func TestCommClient_VideoInformation(t *testing.T) {
	info, err := testCommClient.VideoGetInfo(207511956)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Logf(
		"bvid: %s,title: %s,duration: %d,ownerName: %s,staff: %v,pubdate: %d,noreprint: %d,view: %d,desc: %s,title: %s",
		info.BVID,
		info.Title,
		info.Duration,
		info.Owner.Name,
		info.Staff,
		info.Pubdate,
		info.Rights.NoReprint,
		info.Stat.View,
		info.DescV2[0].RawText,
		info.Pages[0].Part,
	)
}
func TestCommClient_VideoDescription(t *testing.T) {
	desc, err := testCommClient.VideoGetDescription(759949922)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Logf("desc: %s", desc)
}
func TestCommClient_VideoPageList(t *testing.T) {
	list, err := testCommClient.VideoGetPageList(759949922)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	for _, page := range list {
		t.Logf("pageID: %d,title: %s,duration: %d,cid: %d", page.Page, page.Part, page.Duration, page.CID)
	}
}
func TestCommClient_VideoOnlineNum(t *testing.T) {
	total, web, err := testCommClient.VideoGetOnlineNum(759949922, 392402545)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Logf("total: %s,web: %s", total, web)
}
func TestCommClient_VideoTags(t *testing.T) {
	tags, err := testCommClient.VideoTags(759949922)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	for _, tag := range tags {
		t.Logf("name: %s,use: %d,short_content: %s,cover: %s,Isatten: %d", tag.TagName, tag.Count.Use, tag.ShortContent, tag.Cover, tag.IsAtten)
	}
}
func TestCommClient_VideoRecommend(t *testing.T) {
	videos, err := testCommClient.VideoGetRecommend(759949922)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	for _, v := range videos {
		t.Logf("aid: %d,duration: %d,title: %s", v.AID, v.Duration, v.Title)
	}
}
func TestCommClient_VideoShot(t *testing.T) {
	shot, err := testCommClient.VideoShot(759949922, 0, true)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Logf("img_x_len: %d,img_y_len: %d,img_x_size: %d,img_y_size: %d", shot.ImgXLen, shot.ImgYLen, shot.ImgXSize, shot.ImgYSize)
	t.Logf("url: %s", shot.Pvdata)
	t.Logf("image: %v", shot.Image)
	t.Logf("index: %v", shot.Index)
}
func TestCommClient_VideoShot2(t *testing.T) {
	shot, err := testCommClient.VideoShot(759949922, 392402545, false)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Logf("img_x_len: %d,img_y_len: %d,img_x_size: %d,img_y_size: %d", shot.ImgXLen, shot.ImgYLen, shot.ImgXSize, shot.ImgYSize)
	t.Logf("url: %s", shot.Pvdata)
	t.Logf("image: %v", shot.Image)
	// index传入false 则Index属性为空
	t.Logf("index: %v", shot.Index)
}
func TestCommClient_GetUnixNow(t *testing.T) {
	unix, err := testCommClient.GetUnixNow()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Logf("time: %d", unix)
}
func TestCommClient_DanmakuGetLikes(t *testing.T) {
	result, err := testCommClient.DanmakuGetLikes(397011525, []uint64{54109805459813888, 54109892081901568})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	for _, r := range result {
		t.Logf("likes: %d,isLiked: %d,dmid: %s", r.Likes, r.UserLike, r.IDStr)
	}
}
func TestCommClient_GetDailyNum(t *testing.T) {
	result, err := testCommClient.GetDailyNum()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	for k, v := range result {
		t.Logf("%d %d", k, v)
	}
}
func TestCommClient_SpaceTopArchiveGet(t *testing.T) {
	top, err := testCommClient.SpaceGetTopArchive(546195)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Logf("aid: %d,title: %s,reason: %s,inter_video: %t,owner: %s", top.AID, top.Title, top.Reason, top.InterVideo, top.Owner.Name)
}
func TestCommClient_SpaceMasterpiecesGet(t *testing.T) {
	mp, err := testCommClient.SpaceGetMasterpieces(546195)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	for _, archive := range mp {
		t.Logf(
			"aid: %d,title: %s,reason: %s,inter_video: %t,owner: %s,view: %d,height: %d,noreprint: %d",
			archive.AID,
			archive.Title,
			archive.Reason,
			archive.InterVideo,
			archive.Owner.Name,
			archive.Stat.View,
			archive.Dimension.Height,
			archive.Rights.NoReprint,
		)

	}

}
func TestCommClient_SpaceTagsGet(t *testing.T) {
	tags, err := testCommClient.SpaceGetTags(53456)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Log(tags)
}
func TestCommClient_SpaceNoticeGet(t *testing.T) {
	notice, err := testCommClient.SpaceGetNotice(53456)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Log(notice)
}
func TestCommClient_SpaceLastPlayGame(t *testing.T) {
	games, err := testCommClient.SpaceGetLastPlayGame(2)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	for _, game := range games {
		t.Logf("name: %s,image: %s,website: %s", game.Name, game.Image, game.Website)
	}
}
func TestCommClient_DanmakuPBGet(t *testing.T) {
	r, err := testCommClient.DanmakuGetByPb(1, 1176840, 1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Logf("num: %d", len(r.Danmaku))
	for _, dm := range r.Danmaku {
		t.Logf("content: %s,midhash: %s,progress: %d,id: %s", dm.Content, dm.MidHash, dm.Progress, dm.IDStr)
	}
}
func TestCommClient_DanmakuShotGet(t *testing.T) {
	r, err := testCommClient.DanmakuGetShot(759949922)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Log(r)
}
func TestCommClient_SpaceLastVideoCoin(t *testing.T) {
	r, err := testCommClient.SpaceGetLastVideoCoin(7190459)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	for _, i := range r {
		t.Logf("title: %s,duration: %d,owner: %s", i.Title, i.Duration, i.Owner.Name)
	}
}
func TestCommClient_SpaceVideoSearch(t *testing.T) {
	r, err := testCommClient.SpaceSearchVideo(546195, "click", 0, "", 1, 10)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	for _, list := range r.List.Vlist {
		t.Logf("title: %s,view: %d,duration: %s", list.Title, list.Play, list.Length)
	}
	for k, v := range r.List.Tlist {
		t.Logf("tid: %s,name: %s,count: %d", k, v.Name, v.Count)
	}
	t.Logf("videos: %d,pn: %d,ps: %d", r.Page.Count, r.Page.PN, r.Page.PS)
	t.Logf("text: %s,url: %s", r.EpisodicButton.Text, r.EpisodicButton.Uri)
}
func TestCommClient_SpaceChannelList(t *testing.T) {
	list, err := testCommClient.ChanGet(546195)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	for _, c := range list.List {
		t.Logf("chanID: %d,mid: %d,name: %s,intro: %s,cover: %s", c.CID, c.MID, c.Name, c.Intro, c.Cover)
	}
}
func TestCommClient_ChannelVideoGet(t *testing.T) {
	videos, err := testCommClient.ChanGetVideo(546195, 21837, 1, 8)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	for _, v := range videos.List.Archives {
		t.Logf("view: %d,title: %s,ctime: %d", v.Stat.View, v.Title, v.Ctime)
	}
	t.Logf("total: %d,count:%d,name: %s,intro: %s", videos.Page.Count, videos.List.Count, videos.List.Name, videos.List.Intro)
}
func TestCommClient_FavoritesList(t *testing.T) {
	list, err := testCommClient.FavGet(451863725)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	for _, info := range list.List {
		t.Logf("mlid: %d,fid: %d,title: %s,count: %d", info.ID, info.FID, info.Title, info.MediaCount)
	}
}
func TestCommClient_FavoritesDetailGet(t *testing.T) {
	d, err := testCommClient.FavGetDetail(25422594)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Logf(
		"mlid: %d,title: %s,intro: %s,count: %d,ctime: %d,play: %d,owner: %s",
		d.ID,
		d.Title,
		d.Intro,
		d.MediaCount,
		d.Ctime,
		d.CntInfo.Play,
		d.Upper.Name,
	)
}
func TestCommClient_FavoritesResGet(t *testing.T) {
	results, err := testCommClient.FavGetRes(1223365625)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	for _, r := range results {
		t.Logf("id: %d,type: %d", r.ID, r.Type)
	}
}
func TestCommClient_FavoritesResDetail(t *testing.T) {
	d, err := testCommClient.FavGetResDetail(1052622027, 0, "", "", 0, 2, 20)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	for _, m := range d.Medias {
		t.Logf("title: %s,view: %d,favTime: %d", m.Title, m.CntInfo.Play, m.FavTime)
	}
	t.Logf("title: %s,count: %d,owner: %s", d.Info.Title, d.Info.MediaCount, d.Info.Upper.Name)
}
func TestCommClient_VideoPlayURLGet(t *testing.T) {
	// flv
	r, err := testCommClient.VideoGetPlayURL(99999999, 171776208, 112, 128)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Logf("qn: %d,duration: %d", r.Quality, r.TimeLength)
	t.Logf("acceptDesc: %v", r.AcceptDescription)
	for _, u := range r.DURL {
		t.Logf("order: %d,size: %d,url: %s", u.Order, u.Size, u.URL)
	}
}
func TestCommClient_VideoPlayURLGet2(t *testing.T) {
	// dash+HDR
	r, err := testCommClient.VideoGetPlayURL(969628065, 244954665, 112, 16|64)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Logf("qn: %d,duration: %d", r.Quality, r.TimeLength)
	t.Logf("acceptDesc: %v", r.AcceptDescription)
	for _, v := range r.Dash.Video {
		t.Logf("id: %d,codecs: %s,baseURL: %s", v.ID, v.Codecs, v.BaseURL)
	}
	for _, a := range r.Dash.Audio {
		t.Logf("id: %d,codecs: %s,baseURL: %s", a.ID, a.Codecs, a.BaseURL)
	}
}
func TestCommClient_EmoteFreePackageGet(t *testing.T) {
	packs, err := testCommClient.EmoteGetFreePack("reply")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	for _, pack := range packs {
		t.Logf("pack id: %d,text: %s,url: %s", pack.ID, pack.Text, pack.URL)
		for i := 0; i < 5; i++ {
			t.Logf("\temote id: %d,text: %s", pack.Emote[i].ID, pack.Emote[i].Text)
		}
	}
}
func TestCommClient_EmotePackageDetail(t *testing.T) {
	packs, err := testCommClient.EmoteGetPackDetail("reply", []int64{1, 2, 93})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	for _, pack := range packs {
		t.Logf("pack id: %d,text: %s,url: %s", pack.ID, pack.Text, pack.URL)
		for i := 0; i < 5; i++ {
			t.Logf("\temote id: %d,text: %s", pack.Emote[i].ID, pack.Emote[i].Text)
		}
	}
}
func TestCommClient_AudioInformation(t *testing.T) {
	info, err := testCommClient.AudioGetInfo(2445151)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Logf("auid: %d,relatedAID: %d,duration: %d", info.ID, info.AID, info.Duration)
	t.Logf("author: %s,uname: %s,coin: %d,passTime: %d", info.Author, info.Uname, info.CoinNum, info.PassTime)
	t.Logf("play: %d,share: %d,collect: %d,comment: %d,coin: %d", info.Statistic.Play, info.Statistic.Share, info.Statistic.Collect, info.Statistic.Comment, info.CoinNum)
}
func TestCommClient_AudioTags(t *testing.T) {
	tags, err := testCommClient.AudioGetTags(2445151)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	for _, tag := range tags {
		t.Logf("key: %d,info: %s,type: %s,subType: %d", tag.Key, tag.Info, tag.Type, tag.Subtype)
	}
}
func TestCommClient_AudioMembers(t *testing.T) {
	members, err := testCommClient.AudioGetMembers(815861)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	for _, m := range members {
		t.Logf("type: %d", m.Type)
		for _, l := range m.List {
			t.Logf("\tname: %s,memberID: %d", l.Name, l.MemberID)
		}
	}
}
func TestCommClient_AudioLyric(t *testing.T) {
	lrc, err := testCommClient.AudioGetLyric(15664)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Log(lrc)
}
func TestCommClient_AudioStat(t *testing.T) {
	stat, err := testCommClient.AudioGetStat(15664)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Logf("sid: %d,play: %d,comment: %d,collect: %d,share: %d", stat.SID, stat.Play, stat.Comment, stat.Collect, stat.Share)
}
func TestCommClient_AudioPlayURLGet(t *testing.T) {
	data, err := testCommClient.AudioGetPlayURL(2478206, 3)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Logf("auid: %d,type: %d,size: %d", data.SID, data.Type, data.Size)
	t.Logf("timeout: %d,title: %s,cover: %s", data.Timeout, data.Title, data.Cover)
	for _, u := range data.CDNs {
		t.Logf("\turl: %s", u)
	}
	for _, q := range data.Qualities {
		t.Logf("type: %d,size: %d,bps: %s,desc: %s,tag: %s,require: %d,requireDesc: %s", q.Type, q.Size, q.Bps, q.Desc, q.Tag, q.Require, q.RequireDesc)
	}
}
func TestCommClient_ChargeSpaceList(t *testing.T) {
	list, err := testCommClient.ChargeSpaceGetList(546195)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Logf("count: %d, total: %d,display: %d", list.Count, list.TotalCount, list.DisplayNum)
	for _, i := range list.List {
		t.Logf("mid: %d,rank: %d,uname: %s,msg: %s", i.PayMID, i.Rank, i.Uname, i.Message)
	}
}
func TestCommClient_ChargeVideoList(t *testing.T) {
	list, err := testCommClient.ChargeVideoGetList(546195, 250531882)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Logf("count: %d,total: %d,diplay: %d,avCount: %d", list.Count, list.TotalCount, list.DisplayNum, list.AvCount)
	t.Logf("list:")
	for _, i := range list.List {
		t.Logf("\tmid: %d,rank: %d,uname: %s,msg: %s", i.PayMID, i.Rank, i.Uname, i.Message)
	}
}
func TestCommClient_LiveRoomInfo(t *testing.T) {
	info, err := testCommClient.LiveGetRoomInfo(282994)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Logf("roomID: %d,roomStatus: %d,title: %s", info.RoomID, info.RoomStatus, info.Title)
	t.Logf("liveStatus: %d,online: %d,roundStatus: %d", info.LiveStatus, info.Online, info.RoundStatus)
	t.Logf("cover: %s", info.Cover)
	t.Logf("url: %s", info.URL)
}
