//---------------------------------------------------------------------------

#include <System.hpp>
#pragma hdrstop

#include "qrzcom.h"
#include "Main.h"
extern AnsiString qrz;
extern AnsiString keyqrz;
#pragma package(smart_init)
//---------------------------------------------------------------------------

//   Important: Methods and properties of objects in VCL can only be
//   used in a method called using Synchronize, for example:
//
//      Synchronize(&UpdateCaption);
//
//   where UpdateCaption could look like:
//
//      void __fastcall qrzcom::UpdateCaption()
//      {
//        Form1->Caption = "Updated in a thread";
//      }
//--------------------------------------------------------------------------
__fastcall qrzcom::qrzname(AnsiString datos)
{
AnsiString name, surname, addr, pais;
int inic, fina;
inic = datos.Pos("<fname>");
fina = datos.Pos("</fname>");
name = datos.SubString(inic+7,fina-inic-7);
inic = datos.Pos("<name>");
fina = datos.Pos("</name>");
surname = datos.SubString(inic+6,fina-inic-6);
inic = datos.Pos("<addr2>");
fina = datos.Pos("</addr2>");
addr = datos.SubString(inic+7,fina-inic-7);
inic = datos.Pos("<country>");
fina = datos.Pos("</country>");
pais = datos.SubString(inic+9,fina-inic-9);
Mmsstv->HisName->Text = name+" "+surname;
if (pais.Length()>0) {
	Mmsstv->HisQTH->Text = addr+" ("+pais+")";
}else Mmsstv->HisQTH->Text = addr;

}
//---------------------------------------------------------------------------

__fastcall qrzcom::qrzcom(bool CreateSuspended)
	: TThread(CreateSuspended)
{
}
//---------------------------------------------------------------------------
void __fastcall qrzcom::Execute()
{
FreeOnTerminate = true;
	if (qrz=="1") {
		Mmsstv->Memo3->Clear();
		try {
			Mmsstv->NetHTTPClient2->Get("http://xmldata.qrz.com/xml/current/?s="+keyqrz+";callsign="+Mmsstv->HisCall->Text);
			qrzname(Mmsstv->Memo3->Text);
		} catch (...) {

		}

	}
}
//---------------------------------------------------------------------------
