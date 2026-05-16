object frecuencia: Tfrecuencia
  Left = 227
  Top = 108
  BorderStyle = bsDialog
  Caption = 'Frecuencia'
  ClientHeight = 117
  ClientWidth = 171
  Color = clBtnFace
  ParentFont = True
  OldCreateOrder = True
  Position = poMainFormCenter
  Scaled = False
  OnShow = FormShow
  PixelsPerInch = 96
  TextHeight = 13
  object Label1: TLabel
    Left = 8
    Top = 21
    Width = 139
    Height = 13
    Caption = 'Introduce frecuencia en Khz:'
  end
  object Edit1: TEdit
    Left = 8
    Top = 40
    Width = 147
    Height = 27
    Font.Charset = DEFAULT_CHARSET
    Font.Color = clWindowText
    Font.Height = -16
    Font.Name = 'Tahoma'
    Font.Style = []
    NumbersOnly = True
    ParentFont = False
    TabOrder = 0
    OnKeyDown = Edit1KeyDown
  end
  object Button1: TButton
    Left = 80
    Top = 84
    Width = 75
    Height = 25
    Caption = 'Aceptar'
    ModalResult = 1
    TabOrder = 1
    TabStop = False
    OnClick = Button1Click
  end
end
