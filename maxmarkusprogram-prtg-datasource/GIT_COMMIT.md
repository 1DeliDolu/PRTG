### …or create a new repository on the command line

```
echo "# grafana" >> README.md
git init
git add README.md
git commit -m "first commit"
git branch -M main
git remote add origin https://github.com/1DeliDolu/PRTG.git
git push -u origin main
```

### …or push an existing repository from the command line

```
git remote add origin https://github.com/1DeliDolu/PRTG.git
git branch -M main
git push -u origin main
```


## Create a release tag

A tag with the format `vX.X.X` is used to trigger the release workflow. Typically all of your changes will be merged into `main`, and the tag is applied to `main`

```bash
git checkout main
git pull origin main
git tag v1.0.0
git push origin v1.0.0
```


If you need to re-tag the release, the current tag can be removed with these commands:

```bash
git tag -d v1.0.0
git push --delete origin v1.0.0
git checkout main
git pull origin main
```
