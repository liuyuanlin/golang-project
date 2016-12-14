//
//  UserData.h
//  Robot
//
//  Created by PU on 13-4-17.
//  Copyright (c) 2013年 PU. All rights reserved.
//

#ifndef __Robot__UserData__
#define __Robot__UserData__

#include <iostream>
#include <vector>

using namespace std;

//令旗信息
struct TrophyInfo{
    unsigned int        uTownHallLevel;
    unsigned int        uBaseNumber;
    unsigned int        uMaxNumber;
};

class RUserData
{
public:
    RUserData();
    static RUserData* Instance();
    void Init();
    void ReadRobotConfigXML();
    void ReadRobotTrophyXML();
    void ReadRobotToChatXML();
    void SetUDID(string udid);
    string GetUDID(unsigned int index);
    string GetName();
    
    //add by wyc 2013-12-23
    unsigned int GetVillageCount(void) { return m_VillageCount;};
    
    bool IsJustForChat(void) {return m_bJustForChat;};
    bool IsOpenBuild(void) {return m_bOpenBuild;};
    
    string m_Path;
    vector<string> m_UDIDVec;
    vector<string> m_RobotNameVec;
    
    unsigned int m_MaxRobotCount;
    string m_IP;
    int    m_Port;
    unsigned int m_VillageCount;
    unsigned int m_NameIndex;
    int    m_ChannelID;
    
    //add by wyc 2013-12-25
    vector<int>             m_vecPort;              //
    vector<int>             m_vecChannel;
    vector<string>          m_vecIp;                //机器人连接的服务器ip
    vector<unsigned int>    m_vecRobotCount;        //该连接的机器人数量
    vector<unsigned int>    m_vecLevel;             //该村庄的本数（townhall等级）
    
    unsigned int            m_uConfigCount;         //压入机器人配置条数
    //unsigned int            m_uPushOnceCount;       //一次压入机器人数量，目前最大为10
    vector<TrophyInfo*>     m_vecTrophyInfo;
    
    //for test chat server
    vector<string>          m_vecChatIp;
    vector<int>             m_vecChatPort;
    vector<unsigned int>    m_vecChatRobotCount;    
    unsigned int            m_uChatCount;           //chat配置数量
    bool                    m_bJustForChat;
    bool                    m_bOpenFight;
    bool                    m_bOpenBuild;
};

#endif /* defined(__Robot__UserData__) */
