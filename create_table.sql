CREATE TABLE [dbo].[GF_VerificationResults] (
    [Id]                     INT           IDENTITY (1, 1) NOT NULL,
    [Version]                NVARCHAR(10)  NULL,
    [FileType]               NVARCHAR(10)  NULL,
    [DataDate]               DATE          NULL,
    [FileName]               NVARCHAR(255) NULL,
    [RegionCode]             NVARCHAR(10)  NULL,
    [Period]                 NVARCHAR(6)   NULL,
    [RecordsCount]           INT           NULL,
    
    -- Данные записи ZAP
    [DN_Patient_Id]          NVARCHAR(50)  NOT NULL,
    [ENP]                    NVARCHAR(16)  NULL,
    [GenderCode]             INT           NULL,
    [BirthDate]              DATE          NULL,
    [SMO_Code]               NVARCHAR(10)  NULL,
    [Attach_MCode]           NVARCHAR(20)  NULL,
    [Attach_Date]            DATE          NULL,
    [SMO_Region_Code]        NVARCHAR(10)  NULL,
    [Group_RH_Code]          INT           NULL,
    [Group_RH_DS]            NVARCHAR(10)  NULL,
    [DN_Prvs]                INT           NULL,
    [Group_RH_Profile]       NVARCHAR(500) NULL,
    [Group_RH_Name]          NVARCHAR(1000) NULL,
    [DN_Rule_In_Name]        NVARCHAR(1000) NULL,
    
    -- Данные ГИС ОМС о медпомощи (DN_GIS)
    [Trigger_Schet_Filename] NVARCHAR(255) NULL,
    [Trigger_Schet_Code]     NVARCHAR(20)  NULL,
    [Trigger_Nschet]         NVARCHAR(50)  NULL,
    [Trigger_Dschet]         DATE          NULL,
    [Trigger_Idcase]         NVARCHAR(50)  NULL,
    [Trigger_SL_Id]          NVARCHAR(50)  NULL,
    [Trigger_SL_Nhistory]    NVARCHAR(50)  NULL,
    [Trigger_DS_CD]          NVARCHAR(10)  NULL,
    [Trigger_MCode]          NVARCHAR(20)  NULL,
    [Trigger_DT]             DATE          NULL,
    
    -- Результат сверки списка ЗЛ на ДН (DN_LIST)
    [DN_List_Period_CD]      INT           NULL,
    [DN_List_Filename]       NVARCHAR(255) NULL,
    [CODE_L]                 NVARCHAR(50)  NULL,
    [DN_List_Result_Code]    NVARCHAR(10)  NULL,
    [DN_List_Date_Checking]  DATE          NULL,
    [DN_List_Result_Descr]   NVARCHAR(500) NULL,
    
    -- Результат сверки план-графика (DN_PLAN)
    [DN_Plan_Period]         DATE          NULL,
    [DN_Plan_Filename]       NVARCHAR(255) NULL,
    [CODE_P]                 NVARCHAR(50)  NULL,
    [DN_Plan_Result_Code]    INT           NULL,
    [DN_Plan_Date_Checking]  DATE          NULL,
    [DN_Plan_Result_Descr]   NVARCHAR(500) NULL,
    
    -- Служебные поля
    [Insert_DTTM]            DATETIME2     DEFAULT GETDATE(),
    [Update_DTTM]            DATETIME2     DEFAULT GETDATE(),
    [Processed_Date]         DATETIME2     DEFAULT GETDATE(),
    
    CONSTRAINT [PK_GF_VerificationResults] PRIMARY KEY CLUSTERED ([Id] ASC)
);

CREATE NONCLUSTERED INDEX [IX_GF_VerificationResults_DN_Patient_Id] 
    ON [dbo].[GF_VerificationResults] ([DN_Patient_Id]);

CREATE NONCLUSTERED INDEX [IX_GF_VerificationResults_ENP] 
    ON [dbo].[GF_VerificationResults] ([ENP]);

CREATE NONCLUSTERED INDEX [IX_GF_VerificationResults_Processed_Date] 
    ON [dbo].[GF_VerificationResults] ([Processed_Date]);