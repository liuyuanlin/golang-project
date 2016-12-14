//
//  CConnectionMgr.h
//  FightServer
//
//  Created by keenblue on 13-3-30.
//  Copyright (c) 2013年 keenblue. All rights reserved.
//

#ifndef __FightServer__CConnectionMgr__
#define __FightServer__CConnectionMgr__

#include <iostream>


#include "NetWork.h"
#include "INetWork.h"

using namespace Net;

class CFsConnectionMgr
{
public:
    CFsConnectionMgr();
    virtual ~CFsConnectionMgr();
    
    static CFsConnectionMgr* GetFsConnMgr();
    static void Release();
    
    void BeginListen(const char* szIp, uint16 uPort);
    
    bool DispatchEvents();
    
    static void OnAcceptCns(uint32 uId, CAcceptorEx* pAcceptorEx);
    static void OnCnsDisconnected(CAcceptorEx* pAcceptorEx);
    static void OnCnsSomeDataSend(CAcceptorEx* pAcceptorEx);
    static void OnCnsSomeDataRecv(CAcceptorEx* pAcceptorEx);
    static unsigned int m_nSaveFightData;
private:
    
    bool                        m_bQuit;
    static CFsConnectionMgr*    m_pFsConnectionMgr;
    static CNetWork*            m_pNetWork;

};

//保存战斗数据
void SaveFightData(void *Data, uint32 uLength);
void GetSaveDataFileName(std::string &LengthFileName, std::string &DataFileName);

#endif /* defined(__FightServer__CConnectionMgr__) */
