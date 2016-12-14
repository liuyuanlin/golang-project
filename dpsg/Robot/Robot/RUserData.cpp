//
//  UserData.cpp
//  Robot
//
//  Created by PU on 13-4-17.
//  Copyright (c) 2013年 PU. All rights reserved.
//

#include "RUserData.h"

#include <fstream>
#include <stdlib.h>
#include "XMLParser.h"

RUserData::RUserData()
:
m_Path(""), m_NameIndex(0), m_ChannelID(0), m_uConfigCount(0), m_uChatCount(0), m_bJustForChat(false), m_bOpenFight(false), m_bOpenBuild(false)
{}

RUserData* RUserData::Instance()
{
    static RUserData userData;
    return &userData;
}

void RUserData::Init()
{
//    ifstream file((m_Path + "/RobotUDID.txt").c_str());
//    if(!file)return;
    char c[64];
//    while (file.getline(c, 64)){m_UDIDVec.push_back(string(c));}
    
    ifstream fileName((m_Path + "/RobotName.txt").c_str());
    if(!fileName)return;
    while (fileName.getline(c, 64)) {m_RobotNameVec.push_back(string(c));}
    
    ReadRobotConfigXML();
    ReadRobotTrophyXML();
    
    if(m_bJustForChat)
        ReadRobotToChatXML();
    //else
        //std::cout << "fuck" << std::endl;
    
    m_NameIndex = random() % m_RobotNameVec.size();
}

void RUserData::ReadRobotConfigXML()
{
    ifstream file((m_Path + "/RobotConfig.xml").c_str());
    if(!file)
        return;
    
    file.seekg(0, ios::end);
    long long buffLen = file.tellg();
    file.seekg(0, ios::beg);
    char buffer[buffLen];
    file.read(buffer, (int)buffLen);
    
    XMLParser* parser = new XMLParser();
    if(parser->Parse(buffer, (int)buffLen))
    {
        if(parser->SetToFirstChild("config")) do
        {
            //m_MaxRobotCount = parser->GetInt("maxRobot");
            //村庄本数
            m_vecLevel.push_back(parser->GetInt("VillageLevel"));
            //当前本数压入的机器人数量
            m_vecRobotCount.push_back(parser->GetInt("RobotCount"));
            
            //m_IP = parser->GetString("IP");
            m_vecIp.push_back(parser->GetString("IP"));
            m_vecPort.push_back(parser->GetInt("Port"));
            //m_Port = parser->GetInt("Port");
            
            //公用数据---村庄数量和渠道号
            if (parser->HasAttr("VillageCount"))
                m_VillageCount = parser->GetInt("VillageCount");
            if (parser->HasAttr("ChannelID"))
                m_ChannelID = parser->GetInt("ChannelID");
            if (parser->HasAttr("JustForChat"))
            {
                if (1 == parser->GetInt("JustForChat"))
                    m_bJustForChat = true;
                else
                    m_bJustForChat = false;
                //std::cout << "!!!!!!!!!!!!!" << m_bJustForChat << std::endl;
            }
            
            if (parser->HasAttr("OpenFight"))
            {
                if (1 == parser->GetInt("OpenFight"))
                    m_bOpenFight = true;
                else
                    m_bOpenFight = false;
            }
            
            if (parser->HasAttr("OpenBuild"))
            {
                if (1 == parser->GetInt("OpenBuild"))
                    m_bOpenBuild = true;
                else
                    m_bOpenBuild = false;
            }
//            if (parser->HasAttr("PushOnce"))
//                m_uPushOnceCount = parser->GetInt("PushOnce");
            
            ++m_uConfigCount;
        }
        while(parser->SetToNextChild("config"));
    }    
    delete parser;
    file.close();
}

//
void RUserData::ReadRobotTrophyXML()
{
    ifstream file((m_Path + "/RobotTrophy.xml").c_str());
    if(!file)
        return;
    
    file.seekg(0, ios::end);
    long long buffLen = file.tellg();
    file.seekg(0, ios::beg);
    char buffer[buffLen];
    file.read(buffer, (int)buffLen);
    
    XMLParser* parser = new XMLParser();
    if(parser->Parse(buffer, (int)buffLen))
    {
        if(parser->SetToFirstChild("config")) do
        {
            TrophyInfo* pThophy = new TrophyInfo();
            pThophy->uTownHallLevel = parser->GetInt("TownHallLevel");
            pThophy->uBaseNumber = parser->GetInt("BaseNumber");
            pThophy->uMaxNumber = parser->GetInt("MaxNumber");
            
            m_vecTrophyInfo.push_back(pThophy);
        }
        while (parser->SetToNextChild("config"));
    }
    delete parser;
    file.close();
}

//read chat server config
void RUserData::ReadRobotToChatXML()
{
    ifstream file((m_Path + "/RobotToChat.xml").c_str());
    if(!file)
    {
        std::cout << "No such file named RobotToChat.xml" << std::endl;
        return;
    }
    
    file.seekg(0, ios::end);
    long long buffLen = file.tellg();
    file.seekg(0, ios::beg);
    char buffer[buffLen];
    file.read(buffer, (int)buffLen);
    
    XMLParser* parser = new XMLParser();
    if(parser->Parse(buffer, (int)buffLen))
    {
        //unsigned int uRobotCount = 0;
        if(parser->SetToFirstChild("config")) do
        {
            m_vecChatRobotCount.push_back(parser->GetInt("RobotCount"));
            m_vecChatIp.push_back(parser->GetString("IP"));
            m_vecChatPort.push_back(parser->GetInt("Port"));
            //uRobotCount = parser->GetInt("RobotCount");
            ++m_uChatCount;
        }
        while (parser->SetToNextChild("config"));
    }
    std::cout << "连接chat的数量" << m_uChatCount << std::endl;
    delete parser;
    file.close();
}

void RUserData::SetUDID(string udid)
{
    ofstream file((m_Path + "/RobotUDID.txt").c_str(), ios::binary | ios::app);
    if(!file)
        return;
    
    file.seekp(0, ios::end);
    file<<udid.c_str()<<endl;
    file.close();
}

string RUserData::GetUDID(unsigned int index)
{
    return index < m_UDIDVec.size() ? m_UDIDVec.at(index) : "";
}

string RUserData::GetName()
{
    if(m_NameIndex == m_RobotNameVec.size())
        m_NameIndex = 0;
    
    return m_RobotNameVec[m_NameIndex++];
}