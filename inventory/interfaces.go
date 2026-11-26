package main

// A read-only prerequisite to safe driver restoration. Linux device-level
// driver "usb" does not identify which driver owns each interface. These
// observations are neither an enrollment nor an export/restore grant.
// Enumeration is not an atomic hardware snapshot; partial/hotplug cases
// must be rejected by a subsequent independently fenced Fabric transaction.
import (
 "os"
 "path/filepath"
 "regexp"
 "strconv"
 "strings"
)

type Interface struct {
 ID string `json:"id"`
 Configuration int `json:"configuration"`
 Number int `json:"number"`
 AlternateSetting *int `json:"alternate_setting,omitempty"`
 Driver string `json:"driver,omitempty"`
 DriverState string `json:"driver_state"`
 Observation string `json:"observation"`
}

var interfaceID=regexp.MustCompile("^([0-9]+-[0-9]+([.][0-9]+)*):([0-9]+)[.]([0-9]+)$")
var interfaceHex=regexp.MustCompile("^[0-9a-fA-F]{2}$")
func decimalByte(s string)(int,bool){
 if s==""||len(s)>3{return 0,false}
 for _,c:=range s{if c<'0'||c>'9'{return 0,false}}
 n,e:=strconv.Atoi(s)
 return n,e==nil&&n>=0&&n<=255
}
func groupInterfaceEntries(entries []os.DirEntry)map[string][]string{
 out:=make(map[string][]string)
 for _,e:=range entries{
  name:=e.Name()
  if len(name)>80{continue}
  m:=interfaceID.FindStringSubmatch(name)
  if len(m)==5{out[m[1]]=append(out[m[1]],name)}
 }
 return out
}

func observedInterfaceDriver(root,target string)(string,string){
 link:=filepath.Join(target,"driver")
 info,e:=os.Lstat(link)
 if os.IsNotExist(e){return "","UNBOUND"}
 if e!=nil||info.Mode()&os.ModeSymlink==0{return "","UNKNOWN"}
 resolved,e:=filepath.EvalSymlinks(link)
 if e!=nil{return "","UNKNOWN"}
 drivers:=filepath.Join(root,"bus","usb","drivers")
 name:=filepath.Base(resolved)
 if !contained(root,resolved)||filepath.Dir(resolved)!=drivers||!driverName.MatchString(name){
  return "","UNKNOWN"
 }
 st,e:=os.Stat(resolved)
 if e!=nil||!st.IsDir(){return "","UNKNOWN"}
 return name,"BOUND"
}
func observeInterfaces(root,dir,device,id string,names []string)([]Interface,string){
 out:=[]Interface{}
 config,configOK:=decimalByte(attr(device,"bConfigurationValue"))
 count,countOK:=decimalByte(attr(device,"bNumInterfaces"))
 complete:=configOK&&config>0&&countOK&&count==len(names)&&count>0
 seen:=make(map[int]bool)
 for _,name:=range names{
  v:=Interface{ID:name,DriverState:"UNKNOWN",Observation:"PARTIAL"}
  m:=interfaceID.FindStringSubmatch(name)
  if len(m)!=5||m[1]!=id{complete=false;out=append(out,v);continue}
  c,cok:=decimalByte(m[3]);n,nok:=decimalByte(m[4])
  v.Configuration=c;v.Number=n
  if !cok||!nok||strconv.Itoa(c)!=m[3]||strconv.Itoa(n)!=m[4]||
   !configOK||c!=config||seen[n]{
   complete=false;out=append(out,v);continue
  }
  seen[n]=true
  // In Linux sysfs, an interface belongs to the actual device directory.
  // Do not trust a same-looking basename located under a different device.
  target,e:=filepath.EvalSymlinks(filepath.Join(dir,name))
  if e!=nil||!contained(root,target)||filepath.Dir(target)!=device||filepath.Base(target)!=name{
   complete=false;out=append(out,v);continue
  }
  st,e:=os.Stat(target)
  if e!=nil||!st.IsDir(){complete=false;out=append(out,v);continue}
  number:=attr(target,"bInterfaceNumber")
  value,e:=strconv.ParseUint(number,16,8)
  alt,ok:=decimalByte(attr(target,"bAlternateSetting"))
  if !interfaceHex.MatchString(number)||e!=nil||int(value)!=n||!ok{
   complete=false;out=append(out,v);continue
  }
  v.AlternateSetting=&alt
  v.Driver,v.DriverState=observedInterfaceDriver(root,target)
  if v.DriverState=="UNKNOWN"{complete=false;out=append(out,v);continue}
  v.Observation="OBSERVED_NOT_ENROLLED"
  out=append(out,v)
 }
 // Recheck the active configuration and number after enumeration. This
 // detects some topology changes, not all possible ABA or concurrent events.
 if strings.TrimSpace(attr(device,"bConfigurationValue"))!=strconv.Itoa(config){
  complete=false
 }
 again,ok:=decimalByte(attr(device,"bNumInterfaces"))
 if !ok||again!=count{complete=false}
 if complete{return out,"COMPLETE_OBSERVATION_NOT_RESTORE_AUTHORITY"}
 return out,"PARTIAL"
}
