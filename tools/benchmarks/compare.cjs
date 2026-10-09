"use strict";
const fs=require("node:fs"),path=require("node:path"),os=require("node:os"),{execFileSync}=require("node:child_process");
const root=path.resolve(__dirname,"../.."),work=path.resolve(__dirname),out=path.join(root,".git","sdk-benchmark-"+Date.now());
fs.mkdirSync(out,{recursive:true});
function command(exe,args,env){return execFileSync(exe,args,{cwd:work,encoding:"utf8",env:env||process.env,maxBuffer:32<<20,windowsHide:true})}
const environment=JSON.parse(command("go",["env","-json","GOVERSION","GOOS","GOARCH","CGO_ENABLED","GOAMD64"]));
const report={schemaVersion:1,recordedAt:new Date().toISOString(),sdkRevision:command("git",["-C",root,"rev-parse","HEAD"]).trim(),environment,cpu:os.cpus()[0]?.model,officialVersions:{sts:"v2.1.0",openapi:"v2.1.13",tea:"v1.3.13",credentials:"v1.4.5"},method:{workload:"one account-free signed GetCallerIdentity, same response and transport",build:"go build -trimpath; no linker stripping",cold:"unique empty GOCACHE per SDK; includes standard-library compilation; module download excluded",warm:"same command, source and cache, after cold build; output binary removed first",order:["ours","official"],runtime:"serial identity calls; client constructed outside timed loop; 3 x 1s benchmem; no network",limits:"single Windows host; not a network/service benchmark; no statistical generalization"},results:{}};
for(const name of report.method.order){
  const target="./cmd/"+name,env={...process.env,GOCACHE:path.join(out,"cache-"+name)},binary=path.join(out,name+(environment.GOOS==="windows"?".exe":""));
  fs.mkdirSync(env.GOCACHE);
  const packages=command("go",["list","-mod=readonly","-deps","-f","{{.Standard}}|{{.ImportPath}}|{{with .Module}}{{.Path}}|{{.Version}}{{end}}",target]).trim().split(/\r?\n/).map(line=>{const [stdlib,importPath,module,version]=line.split("|");return {stdlib:stdlib==="true",importPath,...(module?{module,version}:{}),project:module?.startsWith("github.com/rambow-cloud/alicloud-go-sdk-x")||false}});
  const modules=[...new Set(packages.filter(p=>p.module&&!p.project).map(p=>p.module+"@"+p.version))].sort();
  const result={dependencies:{totalPackages:packages.length,standardPackages:packages.filter(p=>p.stdlib).length,projectPackages:packages.filter(p=>p.project).length,externalPackages:packages.filter(p=>!p.stdlib&&!p.project).length,externalModules:modules,packages},build:{},runtime:{}};
  for(const temperature of ["cold","warm"]){
    if(fs.existsSync(binary))fs.unlinkSync(binary);
    const start=process.hrtime.bigint();command("go",["build","-mod=readonly","-trimpath","-o",binary,target],env);result.build[temperature+"Milliseconds"]=Number(process.hrtime.bigint()-start)/1e6;
    result.build.binaryBytes=fs.statSync(binary).size;
    if(command(binary,[],env).trim()!=="PASS identity")throw Error("account-free binary validation failed: "+name);
    console.log(name+" "+temperature+" build complete");
  }
  const benchmark=command("go",["test","-mod=readonly","-run","^Example$","-bench","^BenchmarkIdentity$","-benchtime=1s","-count=3","-benchmem",target],env);
  fs.writeFileSync(path.join(out,name+"-benchmark.txt"),benchmark);
  result.runtime.samples=[...benchmark.matchAll(/^BenchmarkIdentity-\d+\s+(\d+)\s+([\d.]+) ns\/op\s+(\d+) B\/op\s+(\d+) allocs\/op/gm)].map(m=>({iterations:Number(m[1]),nanosecondsPerOperation:Number(m[2]),bytesPerOperation:Number(m[3]),allocationsPerOperation:Number(m[4])}));
  if(result.runtime.samples.length!==3)throw Error("unexpected benchmark samples: "+name);
  report.results[name]=result;
}
const filename=path.join(work,"results",environment.GOOS+"-"+environment.GOVERSION+".json");fs.mkdirSync(path.dirname(filename),{recursive:true});fs.writeFileSync(filename,JSON.stringify(report,null,2)+"\n");console.log("Saved "+path.relative(root,filename));
