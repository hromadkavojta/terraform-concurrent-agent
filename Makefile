# main.Url string replace with your app engine Url
# main.Hash string replace with your created secret (used to secure communication)
create:
	go build -ldflags "-X main.Url=https://agent-dot-vojtah-sandbox.ew.r.appspot.com/ -X main.Hash=THISISRANDOMGENERATEDHASHINBUILD" -o tfagent client/main.go
	gcloud app deploy agent/app.yaml
