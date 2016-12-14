#include <iostream>
#include "BaseTickMgr.h"
#include "CConnectionMgr.h"
#include "NetWork.h"
#include "ConfigManager.h"

#include "FileStream.h"
#include "GameLogicDefinitions.h"
#include "LogicInitializer.h"
#include "FightCalculator.h"

#include <fstream>
#include "CFightServerDispatcher.h"

using namespace GameLogic;
using namespace Net;
#define MISSONCONFIG  "misson.config"


vector<string> _splitstring(string input,const char* deli,bool ignoreEmpty = true)
{
	int nend=0;
	int nbegin=0;
	vector<string> outlist;
	while(nend != -1)
	{
		nend = input.find(deli, nbegin);
		if(nend == -1)
		{
			string str = input.substr(nbegin, input.length()-nbegin);
			if(ignoreEmpty)
			{
				if(str != "")
					outlist.push_back(str);
			}
			else
				outlist.push_back(str);
		}
		else
		{
			string str = input.substr(nbegin, nend-nbegin);
			if(ignoreEmpty)
			{
				if(str != "")
					outlist.push_back(str);
			}
			else
				outlist.push_back(str);
		}
		nbegin = nend + strlen(deli);
	}
	return outlist;
}
void ReadConfigFile(string config,string binPath)
{
    GameHub::FileStream file;
    string filename = binPath + "/../designer/" + config;
    file.Open(filename.c_str());
    unsigned int size = file.GetFileSize();
    if (size)
    {
        char* buffer = GH_NEW char[size];
        file.GetBuffer(buffer);
        ConfigManager::Instance()->LoadConfigFromBuffer(config, buffer, size);
        GH_DELETE[] buffer;
        
        //std::cout << "ReadConfigFile : " << config << std::endl;
    }
    else
        std::cout << "ReadConfigFile Faild Path : " << filename << std::endl;
}
void PPE_SleepEx(long ms)
{
#ifdef WIN32
    sleep(ms);
#else
    usleep(ms * 1000);
#endif
}

//debug使用,读取存储的战斗数据回放战斗过程
void ParseFightData()
{
    int i = 0;
    void *pTemp = NULL;
    uint32 uSize = 0;
    char buf[1024*1024] = {};
    std::fstream file1("./fight.length", std::ios::in);
    std::fstream file2("./fight.data", std::ios::in | std::ios::binary);
    while (1) {
        static CFsDispacher oDispacher;
        uint32 uProccessed = 0;
        while (file1 >> uSize) {
            //std::cout << "read data i :" << i << endl;
            //file1 >> uSize;
            std::cout << "uSize: " << uSize << std::endl;
            file2.read(buf, uSize);
            oDispacher.LoopDispatch(buf, uSize, uProccessed, pTemp);
            //PPE_SleepEx(100);
            //break;
        }
        file1.clear();
        file1.seekg(0);
        file2.clear();
        file2.seekg(0);

        if (i > 1000) {
            //break;
        }
        ++i;
    }
    file1.close();
    file2.close();
}

int main(int argc, const char * argv[])
{
    char s[512];
    getcwd(s, 512);
    string binPath = s;
    
    /*
    vector<string> strList = _splitstring(argv[0], "/");
    string binPath;
    for (size_t i=0; i<strList.size()-1; ++i)
    {
        binPath += "/";
        binPath += strList.at(i);
    }*/
    ReadConfigFile(BUILDINGCONFIG,binPath);
    ReadConfigFile(CHARACTERCONFIG,binPath);
    ReadConfigFile(PROJECTILE,binPath);
    ReadConfigFile(MISSONCONFIG,binPath);
    ReadConfigFile(SKILLCONFIG,binPath);
    ReadConfigFile(SPELLSCONFIG,binPath);
    FightCalculator::Instance()->ReadMissionConfigList(MISSONCONFIG);
    LogicInitializer::Instance()->Initialize(false);
    
    std::string strHost = "127.0.0.1";
    uint16 uPort = 8803;
    if(argc > 1)
    {
        //std::cout << "argc " << argc << argv[0] << "," << argv[1] << std::endl;
        strHost = argv[0];
        std::string strPort = argv[1];
        
        std::cout << "input strHost :" << strHost << " : " << strPort << std::endl;
        uPort = atoi(strPort.c_str());
    } else {
        cout << "test!!!!!!!!!!" << endl;
        ParseFightData();
    }
    
    CFsConnectionMgr* pMgr = CFsConnectionMgr::GetFsConnMgr();
    pMgr->BeginListen(strHost.c_str(), uPort);
    
    std::cout << "FightServer Listen to :" << strHost << ":" << uPort << std::endl;
    
    for(;;)
    {
        PPE_SleepEx(1);
        bool bQuit = pMgr->DispatchEvents();
        if(bQuit)
            break;
    }
    
    return 0;
}
