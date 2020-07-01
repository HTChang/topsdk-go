# topsdk-go (淘宝开放平台 SDK in GO)

HTChang/topsdk-go implements the TOP protocol that deals with APIs in single or batch request
with simple json response.
```
go get -u github.com/HTChang/topsdk-go
```
```go
import "github.com/HTChang/topsdk-go"
```
## Reference
[淘宝开放平台API文档](https://open.taobao.com/api.htm)

## Examples

### Create client

```go
client, err := top.NewClient("APP_KEY", "APP_SECRET",
	top.WithSession("SESSION_KEY"))
if err != nil {
    // handle error
    return
}
```

### Single request
```go
res, err := client.DoJson(context.Background(), "taobao.items.onsale.get", top.Parameters{
	"fields":    "total_results,approve_status,num_iid,num",
	"page_no":   pageNo,
	"page_size": pageSize,
})
if tErr, ok := err.(*top.ErrorResponse); ok {
    // can handle the top error for business logic
    return
} else if err != nil {
    // handle other errors 
    return
}

root := res.Get("items_onsale_get_response")
total := root.Get("total_results").MustInt()
itemsBytes, err := root.Get("items").Get("item").MarshalJSON()
// unmarshal to self-defined struct 
// do stuff...
...
```

### Batch request
```go
params := make([]top.Parameters, len(tids))
for i, tid := range tids {
    params[i] = top.Parameters{
        "method": "taobao.trade.fullinfo.get", 
        "tid": tid, 
        "fields": "tid,tid_str,receiver_mobile,receiver_phone",
    }
}
results, err := client.DoJsonBatch(context.Background(), params...)
if tErr, ok := err.(*top.ErrorResponse); ok {
    // can handle the top error for business logic
    return
} else if err != nil {
    // handle other errors
    return
}

for i, res := range results {
    tid := params[i]["tid"].(int64)
    if tErr, ok := res.Err.(*top.ErrorResponse); ok {
        // can handle the top error for business logic 
    	continue
    } else if res.Err != nil { 
        // handle other errors
        continue
    }
    trade := res.Get("trade_fullinfo_get_response").Get("trade")
    mobile := trade.Get("receiver_mobile").MustString()
    phone := trade.Get("receiver_phone").MustString()
    // do stuff...
}
...
```