//----------------------------------------------------------------------------
#ifndef OkCancl2H
#define OkCancl2H
//----------------------------------------------------------------------------
#include <Vcl.ExtCtrls.hpp>
#include <Vcl.Buttons.hpp>
#include <Vcl.StdCtrls.hpp>
#include <Vcl.Controls.hpp>
#include <Vcl.Forms.hpp>
#include <Vcl.Graphics.hpp>
#include <System.Classes.hpp>
#include <System.SysUtils.hpp>
#include <Winapi.Windows.hpp>
#include <System.hpp>
//----------------------------------------------------------------------------
class Tfrecuencia : public TForm
{
__published:
	TEdit *Edit1;
	TLabel *Label1;
	TButton *Button1;
	void __fastcall Button1Click(TObject *Sender);
	void __fastcall FormShow(TObject *Sender);
	void __fastcall Edit1KeyDown(TObject *Sender, WORD &Key, TShiftState Shift);
private:
public:
	virtual __fastcall Tfrecuencia(TComponent* AOwner);
};
//----------------------------------------------------------------------------
extern PACKAGE Tfrecuencia *frecuencia;
//----------------------------------------------------------------------------
#endif    
