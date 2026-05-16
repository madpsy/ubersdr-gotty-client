//---------------------------------------------------------------------------

#ifndef qrzcomH
#define qrzcomH
//---------------------------------------------------------------------------
#include <System.Classes.hpp>
//---------------------------------------------------------------------------
class qrzcom : public TThread
{
private:
protected:
	void __fastcall Execute();
public:
	__fastcall qrzcom(bool CreateSuspended);
	__fastcall qrzname(AnsiString datos);
};
//---------------------------------------------------------------------------
#endif
