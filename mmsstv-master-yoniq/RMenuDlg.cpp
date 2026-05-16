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

#include "RMenuDlg.h"
#include "Main.h"
//--------------------------------------------------------------------- 
#pragma resource "*.dfm"
//TRMenuDialog *RMenuDialog;
static LPCSTR g_help={
"ƒRƒ}ƒ“ƒh\r\n"
"\r\n"
"\\$##...	##=00-FF,  ‘—M‚·‚é•¡”ƒoƒCƒg‚ð16i”‚Å’è‹`‚·‚é\r\n"
"		(—á:\\$FE55AA  -> FEh,55h,AAh)\r\n"
"\\x##	##=00-FF,  ‘—M‚·‚é‚PƒoƒCƒg‚ð16i”‚Å’è‹`‚·‚é\r\n"
"		(—á:\\xFE\\x55\\xAA  -> FEh, 55h, AAh)\r\n"
"\\w##	##=00-99,  ƒfƒBƒŒƒCŽžŠÔ‚ð10ms’PˆÊ‚Ì10i”‚Å’è‹`‚·‚é\r\n"
"		(—á:\\w05  -> wait 50ms)\r\n"
"\\r	ƒLƒƒƒŠƒbƒWƒŠƒ^[ƒ“‚ð‘—M‚·‚é\r\n"
"\\n	ƒ‰ƒCƒ“ƒtƒB[ƒh‚ð‘—M‚·‚é\r\n"
"\\c....	ˆÈ~‚Í‚·‚×‚ÄƒRƒƒ“ƒg\r\n"
"\\\\	'\\'•¶Žš‚ð‘—M‚·‚é\r\n"
"‚»‚Ì‘¼	‚»‚Ì•¶Žš—ñ‚ð‘—M‚·‚é\r\n"
"\r\n"
"##‚Ì‘ã‚í‚è‚Éxx‚ð‹Lq‚·‚é‚ÆA‚»‚Ì•”•ª‚Íxx“ü—Íƒ{ƒbƒNƒX‚Å’è‹`‚µ‚½16i”‚É’u‚«Š·‚í‚è‚Ü‚·B\r\n"
"‚±‚Ì•ÏŠ·‚ÍICOM‚Ì‹@ŽíƒAƒhƒŒƒX‚ðŽw’è‚·‚éê‡‚ÉŽg—p‚µ‚Ü‚·B\r\n"
"		—á:\\$FEFExxE01C0001FD	ixx‚Ì•”•ª‚ª’è‹`“à—e‚É’u‚«Š·‚í‚éj"
};
static LPCSTR g_helpeng={
" Este comando envía un comando de radio a la plataforma. Funcionará para cualquier plataforma. Use la siguiente palabra clave para especificar datos de bytes en número hexadecimal.\r\n"
"\r\n"
"\\$##...		##=00-FF, Especifique la cadena de datos de bytes en formato hexadecimal\r\n"
"			(Example: \\$FE55AA  -> $FE,$55,$AA)\r\n"
"			ICOM CI-V dirección puede ser expresada por xx\r\n"
"\\x##		##=00-FF,Especifique la cadena de datos de bytes en formato hexadecimal\r\n"
" 			(Example: \\xFE\\x55\\xAA  -> $FE, $55, $AA)\r\n"
"\\w##		##=00-99, Especifique el tiempo de retardo\r\n"
"			(Example: \\w05  -> espera 50ms)\r\n"
"\\r		Enviar retorno\r\n"
"\\n		Enviar un avance de línea\r\n"
"\\c....		Comentario\r\n"
"\\\\		'\\' enviar caracter\r\n"
"Otros		esvia caracter sin modificar"
};

//---------------------------------------------------------------------
__fastcall TRMenuDialog::TRMenuDialog(TComponent* AOwner)
	: TForm(AOwner)
{
	FormStyle = ((TForm *)AOwner)->FormStyle;
	Font->Name = ((TForm *)AOwner)->Font->Name;
	Font->Charset = ((TForm *)AOwner)->Font->Charset;

	if( MsgEng ){
		LH->Font->Name = "Arial";
		LH->Font->Charset = ((TForm *)AOwner)->Font->Charset;

		L1->Caption = "Subtítulo";
		L2->Caption = "Comando";
		Caption = "Editar manú";
		SBDel->Caption = "Borrar este menú";
		SBIns->Caption = "Insertar este menú";
		CancelBtn->Caption = "Cancelar";
		LH->Caption = g_helpeng;
    }
    else {
		LH->Caption = g_help;
    }
}
//---------------------------------------------------------------------
void __fastcall TRMenuDialog::UpdateUI(void)
{
	if( SBIns->Visible ){
		SBIns->Enabled = Mmsstv->m_nRadioMenu < RADIOMENUMAX;
    }
}
//---------------------------------------------------------------------
int __fastcall TRMenuDialog::Execute(AnsiString &strTTL, AnsiString &strCMD, int fAdd)
{
	if( fAdd ){
		Caption = MsgEng ? "Agregar menú" : "ƒƒjƒ…[‚Ì’Ç‰Á";
		SBDel->Visible = FALSE;
		SBIns->Visible = FALSE;
    }
	ETTL->Text = strTTL;
    ECMD->Text = strCMD;
    UpdateUI();
    int r = ShowModal();
	strTTL = ETTL->Text;
    strCMD = ECMD->Text;
	return r;
}
//---------------------------------------------------------------------
void __fastcall TRMenuDialog::SBDelClick(TObject *Sender)
{
	ModalResult = 1024;
}
//---------------------------------------------------------------------------
void __fastcall TRMenuDialog::SBInsClick(TObject *Sender)
{
	ModalResult = 1025;
}
//---------------------------------------------------------------------------
void __fastcall TRMenuDialog::SBRefClick(TObject *Sender)
{
	TOpenDialog *pBox = new TOpenDialog(this);
	pBox->Options >> ofCreatePrompt;
	pBox->Options << ofFileMustExist;
	if( MsgEng ){
		pBox->Title = "Abrir archivo";
		pBox->Filter = "Command files(*.rcm)|*.rcm|";
	}
	else {
		pBox->Title = "ƒRƒ}ƒ“ƒh’è‹`ƒtƒ@ƒCƒ‹‚ðŠJ‚­";
		pBox->Filter = "ƒRƒ}ƒ“ƒh’è‹`ƒtƒ@ƒCƒ‹(*.rcm)|*.rcm|";
	}
	pBox->FileName = "";
	pBox->DefaultExt = "rcm";
	pBox->InitialDir = BgnDir;
	if( pBox->Execute() == TRUE ){
		ECMD->Text = pBox->FileName.c_str();
	}
    delete pBox;
}
//---------------------------------------------------------------------------


