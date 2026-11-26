package main

import (
 "os"
 "path/filepath"
 "testing"
)

func addInterface(t *testing.T,root,name,number,alt,driver string){
 t.Helper()
 parent:=filepath.Join(root,"devices","usb1","1-2")
 target:=filepath.Join(parent,name)
 if e:=os.MkdirAll(target,0700);e!=nil{t.Fatal(e)}
 for k,v:=range map[string]string{"bInterfaceNumber":number+"\n","bAlternateSetting":alt+"\n"}{
  if e:=os.WriteFile(filepath.Join(target,k),[]byte(v),0600);e!=nil{t.Fatal(e)}
 }
 if driver!=""{
  p:=filepath.Join(root,"bus","usb","drivers",driver)
  if e:=os.MkdirAll(p,0700);e!=nil{t.Fatal(e)}
  if e:=os.Symlink(p,filepath.Join(target,"driver"));e!=nil{t.Fatal(e)}
 }
 if e:=os.Symlink(target,filepath.Join(root,"bus","usb","devices",name));e!=nil{t.Fatal(e)}
}
func TestInterfaceDriversSeparateAndUnboundObserved(t *testing.T){
 root:=fixture(t)
 dev:=filepath.Join(root,"devices","usb1","1-2")
 if e:=os.WriteFile(filepath.Join(dev,"bNumInterfaces"),[]byte("2\n"),0600);e!=nil{t.Fatal(e)}
 addInterface(t,root,"1-2:1.1","01","2","")
 inv,e:=inventory(root)
 if e!=nil||inv.Status!="OBSERVED"||len(inv.Devices)!=1{t.Fatalf("inventory: %+v %v",inv,e)}
 d:=inv.Devices[0]
 if d.Driver!="usb"||d.InterfaceObservation!="COMPLETE_OBSERVATION_NOT_RESTORE_AUTHORITY"||len(d.Interfaces)!=2{
  t.Fatalf("device/interface conflation: %+v",d)
 }
 a,b:=d.Interfaces[0],d.Interfaces[1]
 if a.Driver!="usbhid"||a.DriverState!="BOUND"||a.Number!=0||
 b.Driver!=""||b.DriverState!="UNBOUND"||b.Number!=1||b.AlternateSetting==nil||*b.AlternateSetting!=2{
  t.Fatalf("invalid observations %+v %+v",a,b)
 }
 t.Log("EDGE_INTERFACE_DRIVER_AND_ALT_OBSERVATION=PASS")
}
func TestInterfaceForeignTargetAndCountMismatch(t *testing.T){
 root:=fixture(t)
 dir:=filepath.Join(root,"bus","usb","devices")
 alias:=filepath.Join(dir,"1-2:1.0")
 if e:=os.Remove(alias);e!=nil{t.Fatal(e)}
 foreign:=filepath.Join(root,"devices","usb1","foreign","1-2:1.0")
 if e:=os.MkdirAll(foreign,0700);e!=nil{t.Fatal(e)}
 if e:=os.Symlink(foreign,alias);e!=nil{t.Fatal(e)}
 inv,e:=inventory(root)
 if e!=nil||inv.Status!="PARTIAL"||inv.Devices[0].Interfaces[0].DriverState!="UNKNOWN"{t.Fatalf("foreign target: %+v %v",inv,e)}
 root=fixture(t)
 dev:=filepath.Join(root,"devices","usb1","1-2")
 if e:=os.WriteFile(filepath.Join(dev,"bNumInterfaces"),[]byte("2\n"),0600);e!=nil{t.Fatal(e)}
 inv,e=inventory(root)
 if e!=nil||inv.Devices[0].InterfaceObservation!="PARTIAL"{t.Fatalf("missing interface accepted: %+v %v",inv,e)}
 t.Log("EDGE_INTERFACE_FOREIGN_AND_INCOMPLETE_DENIED=PASS")
}
func TestInterfaceIdentityAndDriverAmbiguity(t *testing.T){
 root:=fixture(t)
 p:=filepath.Join(root,"devices","usb1","1-2","1-2:1.0")
 if e:=os.WriteFile(filepath.Join(p,"bInterfaceNumber"),[]byte("01\n"),0600);e!=nil{t.Fatal(e)}
 inv,e:=inventory(root)
 if e!=nil||inv.Status!="PARTIAL"{t.Fatalf("interface mismatch: %+v %v",inv,e)}
 root=fixture(t)
 p=filepath.Join(root,"devices","usb1","1-2","1-2:1.0")
 if e:=os.Remove(filepath.Join(p,"driver"));e!=nil{t.Fatal(e)}
 if e:=os.Symlink(filepath.Join(root,"missing-driver"),filepath.Join(p,"driver"));e!=nil{t.Fatal(e)}
 inv,e=inventory(root)
 if e!=nil||inv.Status!="PARTIAL"||inv.Devices[0].Interfaces[0].DriverState!="UNKNOWN"{t.Fatalf("dangling driver accepted: %+v %v",inv,e)}
 t.Log("EDGE_INTERFACE_ID_AND_DRIVER_AMBIGUITY_DENIED=PASS")
}
