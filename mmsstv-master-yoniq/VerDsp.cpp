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



//---------------------------------------------------------------------
#include <vcl.h>
#pragma hdrstop

#include "VerDsp.h"
#include "ComLib.h"
//---------------------------------------------------------------------
#pragma resource "*.dfm"
//TVerDspDlg *VerDspDlg;
//---------------------------------------------------------------------
__fastcall TVerDspDlg::TVerDspDlg(TComponent* AOwner)
	: TForm(AOwner)
{
AnsiString info;
	FormStyle = ((TForm *)AOwner)->FormStyle;
	Font->Name = ((TForm *)AOwner)->Font->Name;
	Font->Charset = ((TForm *)AOwner)->Font->Charset;
	LFree->Font->Name = ((TForm *)AOwner)->Font->Name;
	LFree->Font->Charset = ((TForm *)AOwner)->Font->Charset;
	if( MsgEng ){
		Caption = "Versión Información";
		LFree->Caption = "Free Software";
	}
	Version->Caption = VERTTL;
	LInfo->Caption =
	"I would like to thank my friends in both groups\r\n\r\n"
	"  MM-SSTV@yahoogroups.com\r\n"
	"  mmhamsoft@egroups.co.jp\r\n\r\n"
	"I do not list all those names, for fear that I would forget one.\r\n"
	"And too small space in the window for writing all the name...\r\n\r\n";
	info = "THE NEW CODE CONTAINED IN\n ";
	Label2->Caption = info+VERTTL3;
	Label1->Caption =
	"YONIQ is a modification of version 1.13 by EA1IMW in 2020.\n"
	"YONIQ has been created from the modification of part of the source code provided from Github under GNU Lesser Public General License and the creation of code. In addition, respecting the license, the complete YONIQ code can be downloaded from the original YONIQ download website. Any questions related to YONIQ should be made only through the email yoniq@evl.es.\n"
	"And thank all the colleagues who collaborated in the YONIQ tests.\n"
	"EA1OK, EA3BCX, EA4TM, EA7DUB, EA7JLI, EA7KDR, EA7SL, EB1FE, EA3BE and EA1RCG.\r\n"
	;
}
//---------------------------------------------------------------------
