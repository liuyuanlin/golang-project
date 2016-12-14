//
//  ClientDispacher.h
//  dpsg
//
//  Created by chopdown on 13-1-14.
//
//

#ifndef dpsg_ClientDispacher_h
#define dpsg_ClientDispacher_h

#include "Dispatcher.h"
#include "msg.pb.h"

#define DEFINE_MESSAGE(msg) virtual void OnSync##msg(google::protobuf::Message* m, void* pContext)

class CRobotDispacher:public Net::CMsgDispatcher
{
public:
    CRobotDispacher();
    DEFINE_MESSAGE(LoginResult);
    DEFINE_MESSAGE(PlayerInfo);
    DEFINE_MESSAGE(VillageInfo);
    DEFINE_MESSAGE(PingResult);
    DEFINE_MESSAGE(RpcErrorResponse);
    DEFINE_MESSAGE(MatchPlayer);
    DEFINE_MESSAGE(MatchPlayerResult);
    DEFINE_MESSAGE(LoginCnsInfo);
    DEFINE_MESSAGE(SyncError);
    //add by wyc 2013-12-26
    DEFINE_MESSAGE(UpdatePlayerInfo);
    DEFINE_MESSAGE(Msg);
};

#endif
