//Copyright+LGPL

//-----------------------------------------------------------------------------------------------------------------------------------------------
// Copyright 2020 Eugenio Fdez
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

#include "OKCANCL2.h"
#include "Main.h"
#include "OmniRig_OCX.h"
#include "OmniRig_TLB.h"
//---------------------------------------------------------------------
#pragma resource "*.dfm"
Tfrecuencia *frecuencia;
TOmniRigX *Rig;
//--------------------------------------------------------------------- 
__fastcall Tfrecuencia::Tfrecuencia(TComponent* AOwner)
	: TForm(AOwner)
{
}
//---------------------------------------------------------------------

void __fastcall Tfrecuencia::Button1Click(TObject *Sender)
{
if (Edit1->Text != " ") {
int m_memoria = StrToInt(Edit1->Text);
if (m_memoria>=1810){
	  switch (Mmsstv->memoria)
	  {
		 case 1:
			   Mmsstv->SpeedButton1->Caption = Edit1->Text;
			   Mmsstv->Memo2->Lines->Values["M1"]=Edit1->Text;
			   Mmsstv->Memo2->Lines->SaveToFile("Config.cfg");
			   Close();
			break;
		 case 2:
			   Mmsstv->SpeedButton2->Caption = Edit1->Text;
			   Mmsstv->Memo2->Lines->Values["M2"]=Edit1->Text;
			   Mmsstv->Memo2->Lines->SaveToFile("Config.cfg");
			   Close();
			break;
		 case 3:
			   Mmsstv->SpeedButton3->Caption = Edit1->Text;
			   Mmsstv->Memo2->Lines->Values["M3"]=Edit1->Text;
			   Mmsstv->Memo2->Lines->SaveToFile("Config.cfg");
			   Close();
			break;
		 case 4:
			   Mmsstv->SpeedButton4->Caption = Edit1->Text;
			   Mmsstv->Memo2->Lines->Values["M4"]=Edit1->Text;
			   Mmsstv->Memo2->Lines->SaveToFile("Config.cfg");
			   Close();
			break;
		 case 5:
			   Mmsstv->SpeedButton6->Caption = Edit1->Text;
			   Mmsstv->Memo2->Lines->Values["M5"]=Edit1->Text;
			   Mmsstv->Memo2->Lines->SaveToFile("Config.cfg");
			   Close();
			break;
		 case 6:
			   Mmsstv->SpeedButton7->Caption = Edit1->Text;
			   Mmsstv->Memo2->Lines->Values["M6"]=Edit1->Text;
			   Mmsstv->Memo2->Lines->SaveToFile("Config.cfg");
			   Close();
			break;
		 case 7:
			   Mmsstv->SpeedButton5->Caption = Edit1->Text;
			   Mmsstv->Memo2->Lines->Values["M7"]=Edit1->Text;
			   Mmsstv->Memo2->Lines->SaveToFile("Config.cfg");
			   Close();
			break;
		 case 8:
			   Mmsstv->SpeedButton8->Caption = Edit1->Text;
			   Mmsstv->Memo2->Lines->Values["M8"]=Edit1->Text;
			   Mmsstv->Memo2->Lines->SaveToFile("Config.cfg");
			   Close();
			break;
		 case 9:
			   Mmsstv->SpeedButton9->Caption = Edit1->Text;
			   Mmsstv->Memo2->Lines->Values["M9"]=Edit1->Text;
			   Mmsstv->Memo2->Lines->SaveToFile("Config.cfg");
			   Close();
			break;
		 case 10:
			   Rig = new TOmniRigX(this);
			   Rig->Rig1->set_Freq(m_memoria*1000);
			   delete Rig;
			   Close();
			break;
	  }
}else{
	ShowMessage(L"La frecuencia no corresponde a ninguna banda compatible...");
	}
}else Close();
}
//---------------------------------------------------------------------------


void __fastcall Tfrecuencia::FormShow(TObject *Sender)
{
Edit1->SelectAll();
	  switch (Mmsstv->memoria)
	  {
		 case 11:
			   Edit1->Text = Mmsstv->SpeedButton1->Caption;
			   Mmsstv->memoria = 1;
			   Edit1->SelectAll();
			break;
		 case 12:
			   Edit1->Text = Mmsstv->SpeedButton2->Caption;
			   Mmsstv->memoria = 2;
			   Edit1->SelectAll();
			break;
		 case 13:
			   Edit1->Text = Mmsstv->SpeedButton3->Caption;
			   Mmsstv->memoria = 3;
			   Edit1->SelectAll();
			break;
		 case 14:
			   Edit1->Text = Mmsstv->SpeedButton4->Caption;
			   Mmsstv->memoria = 4;
			   Edit1->SelectAll();
			break;
		 case 15:
			   Edit1->Text = Mmsstv->SpeedButton6->Caption;
			   Mmsstv->memoria = 5;
			   Edit1->SelectAll();
			break;
		 case 16:
			   Edit1->Text = Mmsstv->SpeedButton7->Caption;
			   Mmsstv->memoria = 6;
			   Edit1->SelectAll();
			break;
		 case 17:
			   Edit1->Text = Mmsstv->SpeedButton5->Caption;
			   Mmsstv->memoria = 7;
			   Edit1->SelectAll();
			break;
		 case 18:
			   Edit1->Text = Mmsstv->SpeedButton8->Caption;
			   Mmsstv->memoria = 8;
			   Edit1->SelectAll();
			break;
		 case 19:
			   Edit1->Text = Mmsstv->SpeedButton9->Caption;
			   Mmsstv->memoria = 9;
			   Edit1->SelectAll();
			break;
		 default:
			   Edit1->Text = "";
			   Edit1->SetFocus();
			break;
			}

}
//---------------------------------------------------------------------------






void __fastcall Tfrecuencia::Edit1KeyDown(TObject *Sender, WORD &Key, TShiftState Shift)

{
if (Key == VK_RETURN) {
	Key = VK_INSERT;
	Button1->Click();
}
}
//---------------------------------------------------------------------------


