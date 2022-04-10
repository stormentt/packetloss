packetloss_proto = Proto("packetloss","Packetloss Protocol")

function packetloss_proto.dissector(buffer,pinfo,tree)
  pinfo.cols.protocol = "PACKETLOSS"
  local subtree = tree:add(packetloss_proto,buffer(),"Packetloss Protocol Data")
  subtree:add(buffer(0,32),"HMAC " .. buffer(0,32):bytes():tohex())

  local protobuf_dissector = Dissector.get("protobuf")
  pinfo.private["pb_msg_type"] = "message,packet.Packet"

  pcall(Dissector.call, protobuf_dissector, buffer(32):tvb(), pinfo, subtree)
end

udp_table = DissectorTable.get("udp.port")

udp_table:add(6666,packetloss_proto)
