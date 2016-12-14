//
//  CConnectionMgr.cpp
//  FightServer
//
//  Created by keenblue on 13-3-30.
//  Copyright (c) 2013年 keenblue. All rights reserved.
//

#include "CConnectionMgr.h"
#include "CAcceptorEx.h"
#include "CFightServerDispatcher.h"
#include <fstream>

using namespace std;

CFsConnectionMgr* CFsConnectionMgr::m_pFsConnectionMgr = NULL;
CNetWork* CFsConnectionMgr::m_pNetWork = NULL;
unsigned int CFsConnectionMgr::m_nSaveFightData = 0;

CFsConnectionMgr::CFsConnectionMgr()
:m_bQuit(false)
{
    m_pFsConnectionMgr = this;
    m_pNetWork = &CNetWork::GetSingleton();
    fstream file1("./SaveFightData.config", std::ios::in);
    file1 >> m_nSaveFightData;
    file1.close();
}

CFsConnectionMgr::~CFsConnectionMgr()
{
    m_pFsConnectionMgr = NULL;
    m_pNetWork = NULL;
}

CFsConnectionMgr* CFsConnectionMgr::GetFsConnMgr()
{
    if(m_pFsConnectionMgr == NULL)
    {
        m_pFsConnectionMgr = GH_NEW_T(CFsConnectionMgr, MEMCATEGORY_GENERAL, "");
    }
    
    return m_pFsConnectionMgr;
}

void CFsConnectionMgr::Release()
{
    GH_ASSERT(NULL != m_pFsConnectionMgr);
    GH_DELETE_T(m_pFsConnectionMgr, CFsConnectionMgr, MEMCATEGORY_GENERAL);
}


void CFsConnectionMgr::BeginListen(const char* szIp, uint16 uPort)
{
    m_pNetWork->BeginListen(szIp, uPort, &OnAcceptCns, &OnCnsDisconnected, &OnCnsSomeDataSend, &OnCnsSomeDataRecv);
}

bool CFsConnectionMgr::DispatchEvents()
{
    m_pNetWork->DispatchEvents();
    return m_bQuit;
}

void CFsConnectionMgr::OnAcceptCns(uint32 uId, CAcceptorEx* pAcceptorEx)
{
    std::cout<< "OnAcceptCns : " << uId << std::endl;
}

void CFsConnectionMgr::OnCnsDisconnected(CAcceptorEx* pAcceptorEx)
{
    std::cout<< "OnCnsDisconnected : " << pAcceptorEx->GetId() << std::endl;
    m_pFsConnectionMgr->m_bQuit = true;
}

void CFsConnectionMgr::OnCnsSomeDataSend(CAcceptorEx* pAcceptorEx)
{
}

void CFsConnectionMgr::OnCnsSomeDataRecv(CAcceptorEx* pAcceptorEx)
{
    static CFsDispacher oDispacher;
    uint32 uProccessed=0;
    std::cout << "testconfigfile: " << m_nSaveFightData << " !!!" << std::endl;
    if (m_nSaveFightData == 1) {
        void *pData = pAcceptorEx->GetRecvData();
        uint32 uDataLength = pAcceptorEx->GetRecvDataSize();
        SaveFightData(pData, uDataLength);
        oDispacher.LoopDispatch(pData, uDataLength, uProccessed, pAcceptorEx);
    }else {
        oDispacher.LoopDispatch(pAcceptorEx->GetRecvData(), pAcceptorEx->GetRecvDataSize(), uProccessed, pAcceptorEx);
    }
	pAcceptorEx->PopRecvData(uProccessed);
    
    std::cout<< "OnCnsSomeDataRecv : " << pAcceptorEx->GetId() << std::endl;
}

void SaveFightData(void *pData, uint32 uLength)
{
    string LengthFileName, DataFileName;
    GetSaveDataFileName(LengthFileName, DataFileName);
    std::cout << "getsavefilename: " << LengthFileName << ", " << DataFileName << endl;
    char* pBuf = static_cast<char*>( const_cast<void*>(pData));
    fstream file1(LengthFileName.c_str(), ios::out | ios::app);
    fstream file2(DataFileName.c_str(), ios::out | ios::app | ios::binary);
    
    file1 << uLength << " ";
    file2.write(pBuf, uLength);
    file1.close();
    file2.close();
}

void GetSaveDataFileName(string &LengthFileName, string &DataFileName)
{
    pid_t id = getpid();
    char temp[20];
    sprintf(temp, "%d", id);
    string idString(temp);
    LengthFileName = "./fight" + idString + ".length";
    DataFileName = "./fight" + idString + ".data";
    std::cout << "getsavefilename: " << LengthFileName << ", " << DataFileName << endl;
}
