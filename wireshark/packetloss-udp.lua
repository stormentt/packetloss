-- declare our protocol
packetloss_proto = Proto("packetloss","Packetloss Protocol")
-- create a function to dissect it
function packetloss_proto.dissector(buffer,pinfo,tree)
  pinfo.cols.protocol = "PACKETLOSS"
  local subtree = tree:add(packetloss_proto,buffer(),"Packetloss Protocol Data")
  subtree:add(buffer(0,64),"HMAC " .. buffer(0,64):bytes():tohex())

  local protobuf_dissector = Dissector.get("protobuf")
  pinfo.private["pb_msg_type"] = "message,packet.Packet"

  pcall(Dissector.call, protobuf_dissector, buffer(64):tvb(), pinfo, subtree)
end
-- load the udp.port table
udp_table = DissectorTable.get("udp.port")
-- register our protocol to handle udp port 7777
udp_table:add(6666,packetloss_proto)
