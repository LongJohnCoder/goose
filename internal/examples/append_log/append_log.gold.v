(* autogenerated from append_log *)
From Perennial.go_lang Require Import prelude.

(* disk FFI *)
From Perennial.go_lang Require Import ffi.disk.
Existing Instances disk_op disk_model disk_ty.
Local Coercion Var' (s: string) := Var s.

Module Log.
  Definition S := mkStruct [
    "sz"; "diskSz"
  ].
  Definition T: ty := intT * intT.
  Section fields.
    Context `{ext_ty: ext_types}.
    Definition sz := structF! S "sz".
    Definition diskSz := structF! S "diskSz".
  End fields.
End Log.

Definition writeHdr: val :=
  λ: "log",
    let: "hdr" := NewSlice byteT #4096 in
    UInt64Put "hdr" (Log.sz "log");;
    UInt64Put (SliceSkip #8 "hdr") (Log.sz "log");;
    disk.Write #0 "hdr".

Definition Init: val :=
  λ: "diskSz",
    if: "diskSz" < #1
    then
      (buildStruct Log.S [
         "sz" ::= #0;
         "diskSz" ::= #0
       ], #false)
    else
      let: "log" := buildStruct Log.S [
        "sz" ::= #0;
        "diskSz" ::= "diskSz"
      ] in
      writeHdr "log";;
      ("log", #true).

Definition Get: val :=
  λ: "log" "i",
    let: "sz" := Log.sz "log" in
    if: "i" < "sz"
    then (disk.Read (#1 + "i"), #true)
    else (slice.nil, #false).

Definition writeAll: val :=
  λ: "bks" "off",
    let: "numBks" := slice.len "bks" in
    let: "i" := ref #0 in
    for: (!"i" < "numBks"); ("i" <- !"i" + #1) :=
      let: "bk" := SliceGet "bks" !"i" in
      disk.Write ("off" + !"i") "bk";;
      #true.

Definition Append: val :=
  λ: "log" "bks",
    let: "sz" := Log.sz !"log" in
    if: #1 + "sz" + slice.len "bks" ≥ Log.diskSz !"log"
    then #false
    else
      writeAll "bks" (#1 + "sz");;
      let: "newLog" := buildStruct Log.S [
        "sz" ::= "sz" + slice.len "bks";
        "diskSz" ::= Log.diskSz !"log"
      ] in
      writeHdr "newLog";;
      "log" <- "newLog";;
      #true.

Definition Reset: val :=
  λ: "log",
    let: "newLog" := buildStruct Log.S [
      "sz" ::= #0;
      "diskSz" ::= Log.diskSz !"log"
    ] in
    writeHdr "newLog";;
    "log" <- "newLog".
