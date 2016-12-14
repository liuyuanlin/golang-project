//
//  CFightServerDispatcher.h
//  FightServer
//
//  Created by keenblue on 13-3-30.
//  Copyright (c) 2013年 keenblue. All rights reserved.
//

#ifndef __FightServer__CFightServerDispatcher__
#define __FightServer__CFightServerDispatcher__

#include <iostream>
#include "Dispatcher.h"

//using namespace Net;


#define DEFINE_MESSAGE(msg) virtual void OnSync##msg(google::protobuf::Message* m, void* pContext)


class CFsDispacher :public Net::CMsgDispatcher
{
public:
    
    CFsDispacher();
    ~CFsDispacher();
    
    
    DEFINE_MESSAGE(AttackBegin);
    DEFINE_MESSAGE(AttackEnd);
    DEFINE_MESSAGE(PVEAttackBegin);
};


#endif /* defined(__FightServer__CFightServerDispatcher__) */
