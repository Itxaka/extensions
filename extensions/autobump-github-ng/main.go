package main

import (
	"flag"
	"fmt"
	"github.com/Masterminds/semver/v3"
	pkg "github.com/mudler/luet/pkg/package"
	"github.com/mudler/luet/pkg/tree"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)


func main() {
	var treePath []string
	var revDeps bool
	var treeDir arrayFlags

	flag.BoolVar(&revDeps, "revdeps", false, "Calculate revdeps of the packages.")
	flag.Var(&treeDir, "treedir", "List of dirs in which your packages reside.")
	flag.Parse()

	// cli flags override the env vars
	// no flags? Get from env var
	if len(treeDir) == 0 {
		if v, ok := os.LookupEnv("TREE_DIR"); !ok {
			Error("we require the env var TREE_DIR or the -treedir flag to be set pointing to your packages dir")
		} else {
			treePath = strings.Split(v, " ")
		}
	} else { // flags? get from them
		treePath = strings.Split(treeDir.String(), " ")
	}
	packages := getPackages(treePath, revDeps)
	for _, p := range packages{
		if p.HasLabel("autobump.ignore") && p.GetLabels()["autobump.ignore"] == "true" {
			continue
		}
		p.printLabels()
		// TODO: Parse all the labels here? transform them on the fly here and store them in p while working on this
		strategy := Strategies[p.GetLabels()["autobump.strategy"]]
		version := strategy(p)
		fmt.Printf("\U0001F7E6 Current version %s, remote version %s\n", p.GetVersion(), version)
		currentVersion, err := semver.NewVersion(p.GetVersion())
		if err != nil {
			Error("Error parsing version %s: %s", currentVersion, err)
		}
		remoteVersion, err := semver.NewVersion(version)
		if err != nil {
			Error("Error parsing remote version %s: %s", remoteVersion, err)
		}

		if remoteVersion.Compare(currentVersion) <= 0 {
			fmt.Println("✅ Version is equal or lower, nothing to do")
			continue
		} else {
			fmt.Println("⏫ Remote version is higher, lets bump it!")
			_ = bumpVersion(remoteVersion, p)
		}
	}
}

func bumpVersion(version *semver.Version, p *CustomPackage) error {
	fmt.Printf(p.DefinitionFile)
	file, err := os.Open(p.DefinitionFile)
	if err != nil {
		Error("Error opening def file %s: %s", p.DefinitionFile, err)
	}
	definition, err := ioutil.ReadAll(file)
	if err != nil {
		Error("Error reading def file %s: %s", p.DefinitionFile, err)
	}


	if p.Collection {
		var packages map[string][]pkg.DefaultPackage
		err = yaml.Unmarshal(definition, &packages)
		for index, pack := range packages["packages"] {
			if arePackagesTheSame(pack, p.Package) {
				// set package version
				packages["packages"][index].SetVersion(version.Original())
				// get all lables
				labels := packages["packages"][index].GetLabels()
				// set package version for autobump
				labels["autobump.version"] = version.Original()
				// save back the labels
				packages["packages"][index].Labels = labels
				fmt.Printf("MAP: %v\n", packages)
			}
		}
		//fmt.Printf("collection: %v\n", packages)
	} else {
		var packages pkg.DefaultPackage
		err = yaml.Unmarshal(definition, &packages)
		//fmt.Printf("definition: %v\n", packages)
	}

	//err = yaml.Unmarshal(definition, &packages)
	//if err != nil {
	//	Error("Error unmarshalling def file %s: %s", p.DefinitionFile, err)
	//}
	//fmt.Printf("%v\n", m["version"])
	//m["version"] = version
	//fmt.Printf("%v\n", m["labels"])
	//_, err = yaml.Marshal(m)
	//if err != nil {
	//	Error("Error marshalling data: %s", err)
	//}

	//_, _ = file.Write(toSave)
	//_ = file.Close()
	return nil
}


type arrayFlags []string

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func (i *arrayFlags) String() string { return strings.Join(*i," ") }

// CustomPackage is the same as pkg.Package with an extra var, so we can track the definition file
type CustomPackage struct {
	pkg.Package
	DefinitionFile string
	Collection bool
}

func (c CustomPackage) printLabels()  {
	fmt.Println()
	fmt.Printf("⏳ Checking updates for package: %s/%s\n", c.GetCategory(), c.GetName())
	labels := c.GetLabels()
	for l, v := range labels {
		// clean up the values as it can have \n which pollute our output
		regex, err := regexp.Compile("^$\n")
		if err != nil {
			return
		}
		v = regex.ReplaceAllString(strings.TrimSpace(v), "")
		fmt.Printf("📝 %s ➡ %s\n", l, v)
	}
}


func arePackagesTheSame(p1 pkg.DefaultPackage, p2 pkg.Package) bool {
	// TODO: check also labels???
	if p1.GetPackageName() == p2.GetPackageName() &&
		p1.GetName() == p2.GetName() &&
		p1.GetCategory() == p2.GetCategory() &&
		p1.GetVersion() == p2.GetVersion() {
		return true
	}
	return false
}

// newCustomPackage creates a new custom Package and fills the collection and definition file
func newCustomPackage(p pkg.Package) *CustomPackage{
	customPackage := &CustomPackage{p, "", false}
	if _, ok := os.Stat(fmt.Sprintf("%s/definition.yaml", p.GetPath())); ok == nil {
		customPackage.DefinitionFile = fmt.Sprintf("%s/definition.yaml", p.GetPath())
	} else if _, ok = os.Stat(fmt.Sprintf("%s/collection.yaml", p.GetPath())); ok == nil {
		customPackage.DefinitionFile = fmt.Sprintf("%s/collection.yaml", p.GetPath())
		customPackage.Collection = true
	}
	return customPackage
}

func Error(format string, a ...interface{}) {
	fmt.Println(fmt.Sprintf(format, a...))
	os.Exit(1)
}

// getPackages gets all the packages from the tree dir
func getPackages(treePath []string, revDeps bool) []*CustomPackage {
	var results []*CustomPackage

	reciper := tree.NewInstallerRecipe(pkg.NewInMemoryDatabase(false))

	for _, t := range treePath {
		err := reciper.Load(t)
		if err != nil {
			Error("Error on load tree: %s", err)
		}
	}

	for _, p := range reciper.GetDatabase().World() {
		if revDeps {
			packs, _ := reciper.GetDatabase().GetRevdeps(p)
			for i := range packs {
				revDep := packs[i]
				results = append(results, newCustomPackage(revDep))
			}
		} else {
			results = append(results, newCustomPackage(p))
		}
	}
	return results
}

