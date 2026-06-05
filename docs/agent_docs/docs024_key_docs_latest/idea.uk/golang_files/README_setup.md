page.html has to be in your module directory on the laptop — go:embed reads it at build time, and the build will fail with "no matching files" if it's missing. It's just a copy of the fake-door HTML; grab the page.html I've presented, or if idea_uk_fakedoor.html is already in that folder, cp idea_uk_fakedoor.html page.html. Then:
bash# in the module dir, with page.html present and the updated service.go:
GOPROXY=off GOTOOLCHAIN=local GOOS=linux GOARCH=amd64 go build -o idea .

# atomic swap + restart on the box (no setup.sh needed):
scp idea root@116.203.204.115:/opt/idea/idea.new
ssh root@116.203.204.115 'chmod 755 /opt/idea/idea.new && mv -f /opt/idea/idea.new /opt/idea/idea && systemctl restart idea'
(The mv matters — overwriting a running binary in place gives "text file busy"; writing to a temp name and renaming over it doesn't. That's also what setup.sh's MODE=update does, if you'd rather use that.) Then curl -sS https://idea.uk/ should return the page HTML, and the browser should show the fake-door with the working taster widget.
