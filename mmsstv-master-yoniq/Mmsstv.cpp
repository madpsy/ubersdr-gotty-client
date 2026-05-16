//Copyright+LGPL

//-----------------------------------------------------------------------------------------------------------------------------------------------
// Copyright 2000-2013 Makoto Mori, Nobuyuki Oba
//-----------------------------------------------------------------------------------------------------------------------------------------------
// This file is part of MMSSTV.

// MMSSTV is free software: you can redistribute it and/or modify it under the terms of the GNU Lesser General Public License
// as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

// MMSSTV is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of 
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU Lesser General Public License for more details.

// You should have received a copy of the GNU Lesser General Public License along with MMTTY.  If not, see 
// <http://www.gnu.org/licenses/>.
//-----------------------------------------------------------------------------------------------------------------------------------------------



//---------------------------------------------------------------------------
#include <vcl.h>
#pragma hdrstop
#include <Vcl.Styles.hpp>
#include <Vcl.Themes.hpp>
USEFORM("LogList.cpp", LogListDlg);
USEFORM("LogPic.cpp", LogPicDlg);
USEFORM("LogSet.cpp", LogSetDlg);
USEFORM("MacroKey.cpp", MacroKeyDlg);
USEFORM("Main.cpp", Mmsstv);
USEFORM("MmcgDlg.cpp", MmcgDlgBox);
USEFORM("ListText.cpp", ListTextDlg);
USEFORM("LinearDs.cpp", LinearDspDlg);
USEFORM("LineSet.cpp", LineSetDlg);
USEFORM("Scope.cpp", TScope);
USEFORM("RepSet.cpp", RepSetDlg);
USEFORM("RMenuDlg.cpp", RMenuDialog);
USEFORM("RxView.cpp", RxViewDlg);
USEFORM("StockVew.cpp", StockView);
USEFORM("VerDsp.cpp", VerDspDlg);
USEFORM("ZoomView.cpp", ZoomViewDlg);
USEFORM("TextEdit.cpp", TextEditDlg);
USEFORM("TextIn.cpp", TextInDlg);
USEFORM("TH5Len.cpp", TH5LenDlg);
USEFORM("RadioSet.cpp", RADIOSetDlg);
USEFORM("Option.cpp", OptionDlg);
USEFORM("OKCANCL2.cpp", frecuencia);
USEFORM("PerSpect.cpp", PerSpectDlg);
USEFORM("PlayDlg.cpp", PlayDlgBox);
USEFORM("PrevView.cpp", PrevViewBox);
USEFORM("Qsodlg.cpp", QSODlgBox);
USEFORM("PicFilte.cpp", PicFilterDlg);
USEFORM("PicRect.cpp", PicRectDlg);
USEFORM("PicSel.cpp", PicSelDlg);
USEFORM("HistView.cpp", HistViewDlg);
USEFORM("InputWin.cpp", InputWinDlg);
USEFORM("FreqDisp.cpp", FreqDispDlg);
USEFORM("fileview.cpp", FileViewDlg);
USEFORM("ColorSet.cpp", ColorSetDlg);
USEFORM("BitMask.cpp", BitMaskDlg);
USEFORM("ClockAdj.cpp", ClockAdjDlg);
USEFORM("ColorBar.cpp", ColorBarDlg);
USEFORM("ConvDef.cpp", ConvDefDlg);
USEFORM("ExtCmd.cpp", ExtCmdDlg);
USEFORM("CtrBtn.cpp", CtrBtnWnd);
//---------------------------------------------------------------------------
WINAPI WinMain(HINSTANCE, HINSTANCE, LPSTR, int)
{
	LPSTR pCmd = GetCommandLine();
    if( strstr(pCmd, "-Z") == NULL ){
		HWND hWnd = ::FindWindow("TMmsstv", NULL);
		if( hWnd != NULL ){
			hWnd = ::FindWindow("TAppBuilder", NULL);
			if( hWnd == NULL ){
				return 0;
			}
		}
	}

	try
	{
		Application->Initialize();
		Application->Title = "YONIQ";
		TStyleManager::TrySetStyle("Glossy");
		Application->CreateForm(__classid(TMmsstv), &Mmsstv);
		Application->CreateForm(__classid(Tfrecuencia), &frecuencia);
		Application->Run();
	}
	catch (Exception &exception)
	{
		Application->ShowException(&exception);
	}
	return 0;
}
//---------------------------------------------------------------------------

