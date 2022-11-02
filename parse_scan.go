package evm

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/anaskhan96/soup"
	"github.com/chainlydev/evmparser/lib"
	"github.com/influxdata/influxdb/pkg/slices"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/sync/errgroup"
)

var ScanJobs map[string]bool

type ScanParse struct {
	client *lib.MongoConnect
	jobs   map[string]bool
	ctx    context.Context
}

var scanParse *ScanParse

func NewScanParse() *ScanParse {
	client := lib.NewMongo(nil)
	ctx := context.Background()
	if scanParse == nil {
		scanParse = &ScanParse{client: client, jobs: make(map[string]bool), ctx: ctx}
		scanParse.Worker()
	}
	return scanParse
}

func (tp ScanParse) GetCoin(address string) (string, string, map[string]string, string, []string) {
	var tags []string

	var site string
	var socialData = make(map[string]string)
	var image string
	var typeData string
	resp, err := soup.Get("https://etherscan.io/token/" + address)
	if err == nil {
		doc := soup.HTMLParse(resp)
		offSite := doc.Find("div", "id", "ContentPlaceHolder1_tr_officialsite_1")
		panels := doc.FindAll("div", "class", "h-100")
		if len(panels) > 1 {
			socials := panels[1].FindAll("a", "class", "link-hover-secondary")
			for _, social := range socials {
				sc := strings.Split(social.Attrs()["data-original-title"], ": ")
				socialData[strings.ToLower(strings.TrimSpace(sc[0]))] = strings.TrimSpace(sc[1])
			}
		}
		if offSite.Error == nil {

			if offSite.HTML() != "" {
				site = offSite.Find("a").Text()
			}
		}

		tdata := doc.Find("div", "id", "ContentPlaceHolder1_divSummary")
		if tdata.Error == nil {
			typeData = tdata.Find("span").Text()
		}
		imgd := doc.Find("h1").Find("img")
		if imgd.Error == nil {
			img := imgd.Attrs()
			if img != nil {
				if img["src"] != "" {
					if !strings.Contains(img["src"], "data:image") && !strings.Contains(img["src"], "empty-token.png") {
						image = img["src"]
					}
				}
			}
		}

		doc2 := doc.Find("div", "class", "py-3")
		if doc2.Error == nil {
			resp2 := doc2.FindAll("", "class", "u-label--xs")

			for _, root := range resp2 {
				tags = append(tags, strings.TrimSpace(root.Text()))
			}
			return site, typeData, socialData, image, tags
		}
	}
	return "", "", nil, "", nil
}
func (tp ScanParse) AddressDataParse(doc soup.Root, address string) {
	var tags []string
	var creator string
	var creatorAddr string
	var proxy string
	var tagsM []string
	var socialData = make(map[string]string)
	var image string
	doc2 := doc.Find("div", "class", "py-3")
	imgBase := doc.Find("h1").Find("img")
	if imgBase.Error == nil {
		img := imgBase.Attrs()
		if img != nil {
			if img["src"] != "" {
				if !strings.Contains(img["src"], "data:image") && !strings.Contains(img["src"], "empty-token.png") {
					image = img["src"]
				}
			}
		}
	}

	doc3 := doc.Find("div", "id", "ContentPlaceHolder1_cardright")
	if doc3.Error != nil {
		chall := doc.Find("form", "id", "challenge-form")
		if chall.Error == nil {
			data := url.Values{}
			for _, item := range chall.FindAll("input") {
				params := item.Attrs()
				data.Add(params["name"], params["value"])

			}
			urli := chall.Attrs()["action"]
			resp, _ := soup.PostForm("https://etherscan.io/"+urli, data)
			doc = soup.HTMLParse(resp)
			tp.AddressDataParse(doc, address)
		}
		return
	}
	doc3 = doc3.Find("div", "id", "ContentPlaceHolder1_trContract").Find("a")
	readAsProxyContract := doc.Find("span", "id", "ContentPlaceHolder1_readProxyMessage")
	tokenInfo := doc.Find("div", "id", "ContentPlaceHolder1_tr_tokeninfo")
	var site = ""
	var typeData = ""
	if tokenInfo.Error == nil {
		site, typeData, socialData, image, tagsM = tp.GetCoin(address)
		for _, s := range tagsM {
			tags = append(tags, s)
		}
	}
	_ = site
	_ = typeData
	if readAsProxyContract.Error == nil {
		if readAsProxyContract.Find("a").Error == nil {

			proxy = readAsProxyContract.Find("a").Text()
		}
	}
	creator = doc3.Text()
	creatorAddr = strings.Split(doc3.Attrs()["href"], "/")[2]
	resp2 := doc2.FindAll("", "class", "u-label--xs")

	for _, root := range resp2 {
		if !slices.Exists(tags, strings.TrimSpace(root.Text())) {
			tags = append(tags, strings.TrimSpace(root.Text()))
		}
	}
	if image != "" {

		downloadFile("https://etherscan.io/"+image, "images/"+address+".png")
		image = address + ".png"
	}
	resp, err := tp.client.Collection("token").UpdateMany(context.Background(), bson.M{"address": address}, bson.M{"$set": bson.M{
		"image":            image,
		"tags":             tags,
		"is_porxy":         proxy != "",
		"proxy_address":    proxy,
		"social":           socialData,
		"deployer":         creator,
		"deployer_address": creatorAddr,
	}})
	if err != nil {
		fmt.Println(err, resp)
	}
}
func downloadFile(URL, fileName string) error {
	//Get the response bytes from the url
	response, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil
	}
	//Create a empty file
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	//Write the bytes to the fiel
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}
func (tp ScanParse) GoAddress(address string) {
	resp, err := soup.Get("https://etherscan.io/address/" + address)

	if err == nil {
		doc := soup.HTMLParse(resp)
		tp.AddressDataParse(doc, address)
	}
	delete(tp.jobs, address)

}
func (tp ScanParse) Worker() {
	g, ctx := errgroup.WithContext(tp.ctx)
	t := time.NewTicker(2 * time.Second)
	g.Go(func() error {
		for {
			<-t.C
			caller := ""
			for k, v := range tp.jobs {
				if !v {
					caller = k
					break
				}
			}
			tp.GoAddress(caller)
			select {

			case <-ctx.Done():
				return ctx.Err()
			}

		}
	})

}
func (tp ScanParse) GetAddress(address string) {
	tp.jobs[address] = false
}
