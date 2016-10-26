// Copyright 2011 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
    "fmt"
    "os"
    "encoding/json"
    "sync"
    "os/exec"
    "time"
    "net/url"
    "errors"
    "io/ioutil"

    "strconv"
    "strings"
    "regexp"
)

import (
    "github.com/lxn/walk"
    "github.com/ChimeraCoder/anaconda"
    . "github.com/lxn/walk/declarative"
)

type Foo struct {
    Index   int
    Bar     string
    Baz     string
    Name    string
    Quux    time.Time
    checked bool
}

type FooModel struct {
    walk.TableModelBase
    walk.SorterBase
    sortColumn int
    sortOrder  walk.SortOrder
    items      []*Foo
}

type Animal struct {
    Name          string
}

type Tokens struct{
    Oauth_token_secret string `json:"oauth_token_secret"`
    Oauth_token        string `json:"oauth_token"`
}

func NewFooModel() *FooModel {
    m := new(FooModel)
    //m.ResetRows()
    return m
}

// Called by the TableView from SetModel and every time the model publishes a
// RowsReset event.
func (m *FooModel) RowCount() int {
    return len(m.items)
}



// Called by the TableView to retrieve if a given row is checked.
func (m *FooModel) Checked(row int) bool {
    return m.items[row].checked
}

// Called by the TableView when the user toggled the check box of a given row.
func (m *FooModel) SetChecked(row int, checked bool) error {
    m.items[row].checked = checked

    return nil
}



func (m *FooModel) Len() int {
    return len(m.items)
}

func (m *FooModel) Less(i, j int) bool {
    a, b := m.items[i], m.items[j]

    c := func(ls bool) bool {
        if m.sortOrder == walk.SortAscending {
            return ls
        }

        return !ls
    }

    switch m.sortColumn {
    case 0:
        return c(a.Index < b.Index)

    case 1:
        return c(a.Bar < b.Bar)

    case 2:
        return c(a.Baz < b.Baz)

    case 3:
        return c(a.Quux.After(b.Quux))
    }

    panic("unreachable")
}

func (m *FooModel) Swap(i, j int) {
    m.items[i], m.items[j] = m.items[j], m.items[i]
}


func stringInSlice(a string, list []string) bool {
    for _, b := range list {
        if b == a {
            return true
        }
    }
    return false
}
func prepend(v *Foo, slice []*Foo) []*Foo{
    return append([]*Foo{v}, slice...)
}

var rep *regexp.Regexp
var repMutex *sync.Mutex
func init(){
      rep = regexp.MustCompile(`参戦ID：([a-zA-Z\d]+?)\s*(Lv\d+)\s*(.*)`)
      repMutex=&sync.Mutex{}
      boxcomboMutex=&sync.Mutex{}
}

func ParseItem(tweet anaconda.Tweet)(*Foo,error){

        tmp:= strings.Split(tweet.CreatedAt," ")

        tmptime:=strings.Split(tmp[3],":")
        hoge:=map[string]time.Month{
            "Jan":time.January,
            "Feb":time.February,
            "Mar":time.March,
            "Apr":time.April,
            "May":time.May,
            "Jun":time.June,
            "Jul":time.July,
            "Aug":time.August,
            "Sep":time.September,
            "Oct":time.October,
            "Nov":time.November,
            "Dec":time.December,
        }
        Atoi:=func(str string)int{
            i,_:=strconv.Atoi(str)
            return i;
        }

        t := time.Date(Atoi(tmp[5]), hoge[tmp[1]], Atoi(tmp[2]), Atoi(tmptime[0]), Atoi(tmptime[1]), Atoi(tmptime[2]), 0, time.UTC)



        repMutex.Lock()
        res:=rep.FindAllStringSubmatch(tweet.Text, -1)
        repMutex.Unlock()
        if res==nil||len(res)==0||len(res[0])==0{
            return nil,errors.New("cant match")
        }


        return &Foo{
            Index: 0,//i,
            Bar:   res[0][1],
            Baz:   res[0][2]+" "+res[0][3],
            Name:  res[0][3],
            Quux:  t,
        },nil


}
var boxcomboMutex *sync.Mutex

func getTweetBulk()(anaconda.SearchResponse, error){
    v := url.Values{}
    v.Set("count", "99")
    searchString:="参加者募集！参戦ID："
    fmt.Println(boxcombo.Text())
    if boxcombo.Text() != ""{
        searchString+=" "+boxcombo.Text()
    }
    combos[0]=boxcombo.Text()
    fmt.Println(searchString)
    apilock.Lock()
    ret,err:=api.GetSearch(searchString, v)
    apilock.Unlock()
    return ret,err
}

func (m *FooModel) Value(row, col int) interface{} {
    item := m.items[row]

    switch col {
    case 0:
        return item.Index

    case 1:
        return item.Bar

    case 2:
        return item.Baz

    case 3:
          duration := time.Now().Sub(item.Quux)
        hours0 := int(duration.Hours())
        days := hours0 / 24
        hours := hours0 % 24
        mins := int(duration.Minutes()) % 60
        secs := int(duration.Seconds()) % 60

        daystring:=""
        if days!=0{daystring+=fmt.Sprintf("%d日",days) }
        if days!=0 || hours!=0{daystring+=fmt.Sprintf("%d時間",hours) }
        return daystring+fmt.Sprintf("%d分%d秒前",mins,secs)
    }

    panic("unexpected col")
}

func (m *FooModel) ResetRows() {
    // Create some random data.
    m.items = make([]*Foo, 1)

    m.items[0]= &Foo{
            Index: 0,
            Bar:   "aaa",
            Baz:   "bbb",
            Name:  "ddd",
            Quux:  time.Now(),
    }

    if api==nil{
        return
    }

    searchResult,_:=getTweetBulk()
    m.items= make([]*Foo, 0)


    for _ , tweet := range searchResult.Statuses {
        newItem, err:=ParseItem(tweet)
        if err!=nil{
            continue
        }
        if !stringInSlice(newItem.Name,combos){
            combos=append(combos,newItem.Name)
        }
        m.items=append(m.items,newItem)
    }


	boxcombo.SetModel(combos)
	boxcombo.SetCurrentIndex(0)

    // Notify TableView and other interested parties about the reset.
    m.PublishRowsReset()

    m.Sort(m.sortColumn, m.sortOrder)
}
const (
    configJson = "tokens.ini"
)

var combos []string
var api *anaconda.TwitterApi
var boxcombo *walk.ComboBox
var animal Animal
var apilock sync.Mutex

func startStream(){
    for {
        apilock.Lock()
        if api!=nil{
            apilock.Unlock()
            break;
        }

        apilock.Unlock()
    }

    tmp:=url.Values{}
    str:=""
    for i:=20;i<=200;i+=5{
        str+="Lv"+fmt.Sprintf("%d,",i)
    }
    fmt.Println(str)
    tmp.Set("track",str)
    apilock.Lock()

    fmt.Println("apiからstream発行")
    twitterStream := api.PublicStreamFilter(tmp)
    //twitterStream := api.PublicStreamSample(nil)
    apilock.Unlock()
    for {
        //fmt.Println("streamから取り出そう")
        x := <-twitterStream.C
        //fmt.Println("streamから取り出した")
        switch tweet := x.(type) {
        case anaconda.Tweet:
            fmt.Println(tweet.Text)
            newItem,err:=ParseItem(tweet)
            if err!=nil{
                continue
            }
            if strings.Contains(newItem.Name,boxcombo.Text())==false{
                continue
            }
           if !stringInSlice(newItem.Name,combos){
                combos=append(combos,newItem.Name)
                boxcombo.SetModel(combos)
                boxcombo.SetCurrentIndex(0)
            }
            model.items=prepend(newItem,model.items)
            model.PublishRowsReset()
            //m.Sort(m.sortColumn, m.sortOrder)
            fmt.Println("-----------")
        case anaconda.StatusDeletionNotice:
            // pass
        default:
            fmt.Printf("unknown type(%T) : %v \n", x, x)
        }
    }

}



func setAPIfromJson(){
    if _, err := os.Stat(configJson);err ==nil{
        res,err2:=ioutil.ReadFile(configJson)
        if err2 != nil{
            panic(err2)
        }
        var mt Tokens
        json.Unmarshal(res, &mt)
        apilock.Lock()
        api= anaconda.NewTwitterApi(mt.Oauth_token, mt.Oauth_token_secret)
        apilock.Unlock()
    }
}
var model *FooModel
func main() {
    fmt.Println(os.Getenv("HTTP_PROXY"),os.Getenv("HTTPS_PROXY"))
    anaconda.SetConsumerKey(ConsumerKey)
    anaconda.SetConsumerSecret(ConsumerSecret)

    setAPIfromJson()


    combos=[]string{""}
    var tv *walk.TableView

    mw := &walk.MainWindow{}

    boxcombo=&walk.ComboBox{}
    var db *walk.DataBinder

    model = NewFooModel()

    animal = Animal{}
/*
go func(){
    for true{
        <-time.After(15 *1000 * time.Millisecond)
        model.ResetRows()
    }

}()*/

    go startStream()



    _,err:=MainWindow{
        AssignTo: &mw,
        Title:  "参戦IDさがす君",
        Size:   Size{500, 600},
        Layout: VBox{MarginsZero: true},
        DataBinder: DataBinder{
            AssignTo:       &db,
            DataSource:     animal,
        },
        Children: []Widget{
            Label{Text: "1. Connect to Twitter with your account."},
            PushButton{
                Text: "Connect to Twitter",
                OnClicked: func() {
                    url,credential,err:=anaconda.AuthorizationURL("oob")
                    if err != nil {
                            fmt.Printf("%v", err)
                        }
                    fmt.Println("AuthorizationURL : "+url)
                    exec.Command("cmd", "/C", "start", url,"title" ).Run()

                    if cmd, err := RunAnimalDialog(mw, &animal); err != nil {
                        fmt.Println(err)
                    } else if cmd == walk.DlgCmdOK {
                        fmt.Printf("%v",animal.Name)
                        walk.MsgBox(mw, "Open", "OK!", walk.MsgBoxIconInformation)
                        _,value,err:= anaconda.GetCredentials(credential, animal.Name)
                        if err !=nil{
                            fmt.Println(err)
                        }
                        fmt.Printf("%v",value)
                        apilock.Lock()
                        api= anaconda.NewTwitterApi(value["oauth_token"][0], value["oauth_token_secret"][0])
                        apilock.Unlock()
                        bytes, _ := json.Marshal(Tokens{ value["oauth_token_secret"][0], value["oauth_token"][0]})
                        ioutil.WriteFile(configJson, bytes, os.ModePerm)
                    }

                },
            },
            Label{Text: "2. ロードする"},
            PushButton{
                Text:      "Reset Rows",
                OnClicked: model.ResetRows,
            },
            Label{Text: "3. 種類を絞り込む（絞り込む場合ロードしなおしてください）"},

            ComboBox{
                AssignTo: &boxcombo,
                Editable: true,
            },

            Label{Text: "4. 下から選択してコピー（クリックすれば自動でコピーされます）"},
            TableView{
                AssignTo:              &tv,
                AlternatingRowBGColor: walk.RGB(255, 255, 224),
                CheckBoxes:            false,
                ColumnsOrderable:      true,
                MultiSelection:        false,
                Columns: []TableViewColumn{
                    {Title: "#",Width: 50},
                    {Title: "参戦ID"},
                    {Title: "名前", Alignment: AlignFar, Width: 150},
                    {Title: "時刻"},
                },
                Model: model,
                OnCurrentIndexChanged: func() {
                    fmt.Printf("SelectedIndexes: %v\n", tv.CurrentIndex())
                    if (tv.CurrentIndex())<0{return}
                    if err := walk.Clipboard().SetText(model.items[tv.CurrentIndex()].Bar); err != nil {
                        fmt.Print("Copy: ", err)
                    }
                    fmt.Printf(model.items[tv.CurrentIndex()].Bar)
                },
            },
        },
    }.Run()
    if err!=nil{
        panic(err)
    }

}




func RunAnimalDialog(owner walk.Form, animal *Animal) (int, error) {
    var dlg *walk.Dialog
    var db *walk.DataBinder
    var ep walk.ErrorPresenter
    var acceptPB, cancelPB *walk.PushButton

    return Dialog{
        AssignTo:      &dlg,
        Title:         "Input PIN",
        DefaultButton: &acceptPB,
        CancelButton:  &cancelPB,
        DataBinder: DataBinder{
            AssignTo:       &db,
            DataSource:     animal,
            ErrorPresenter: ErrorPresenterRef{&ep},
        },
        MinSize: Size{300, 300},
        Layout:  VBox{},
        Children: []Widget{
            Composite{
                Layout: HBox{},
                Children: []Widget{
                    Label{
                        Text: "Enter PIN:",
                    },
                    LineEdit{
                        Text: Bind("Name"),
                    },
                },
            },
            Composite{
                Layout: HBox{},
                Children: []Widget{
                    HSpacer{},
                    PushButton{
                        AssignTo: &acceptPB,
                        Text:     "OK",
                        OnClicked: func() {
                            if err := db.Submit(); err != nil {
                                fmt.Print(err)
                                return
                            }

                            dlg.Accept()
                        },
                    },
                    PushButton{
                        AssignTo:  &cancelPB,
                        Text:      "Cancel",
                        OnClicked: func() { dlg.Cancel() },
                    },
                },
            },
        },
    }.Run(owner)
}
