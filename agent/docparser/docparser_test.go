// Copyright 2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not
// use this file except in compliance with the License. A copy of the
// License is located at
//
// http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
// either express or implied. See the License for the specific language governing
// permissions and limitations under the License.

// package parser contains utilities for parsing and encoding MDS/SSM messages.
package docparser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/aws/amazon-ssm-agent/agent/contracts"
	"github.com/aws/amazon-ssm-agent/agent/docmanager/model"
	"github.com/aws/amazon-ssm-agent/agent/jsonutil"
	"github.com/aws/amazon-ssm-agent/agent/log"
	"github.com/stretchr/testify/assert"
)

const (
	testOrchDir    = "test-orchestrationDir"
	testS3Bucket   = "test-s3Bucket"
	testS3Prefix   = "test-s3KeyPrefix"
	testMessageID  = "test-messageID"
	testDocumentID = "test-documentID"
	testWorkingDir = "test-defaultWorkingDirectory"
)
const parameterdocument = `{"schemaVersion":"1.2","description":"","parameters":{"commands":{"type":"StringList"}},"runtimeConfig":{"aws:runPowerShellScript":{"properties":[{"id":"0.aws:runPowerShellScript","runCommand":"{{ commands }}"}]}}}`
const invaliddocument = `{"schemaVersion":"1.2","description":"PowerShell.","FOO":"bar"}`
const testparameters = `{"commands":["date"]}`

var sampleMessageFiles = []string{
	"testdata/sampleMessageVersion2_0.json",
	"testdata/sampleMessage.json",
	"testdata/sampleMessageVersion2_2.json",
}

func TestParseDocument_ValidRuntimeConfig(t *testing.T) {
	mockLog := log.NewMockLog()

	testParserInfo := DocumentParserInfo{
		OrchestrationDir:  testOrchDir,
		S3Bucket:          testS3Bucket,
		S3Prefix:          testS3Prefix,
		MessageId:         testMessageID,
		DocumentId:        testDocumentID,
		DefaultWorkingDir: testWorkingDir,
	}

	validdocumentruntimeconfig := loadFile(t, "../framework/service/runcommand/mds/testdata/validcommand12.json")
	var testDocContent contracts.DocumentContent
	err := json.Unmarshal(validdocumentruntimeconfig, &testDocContent)
	if err != nil {
		assert.Error(t, err, "Error occured when trying to unmarshal validDocument")
	}
	pluginsInfo, err := ParseDocument(mockLog, &testDocContent, testParserInfo, nil)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(pluginsInfo))

	pluginInfoTest := pluginsInfo[0]
	assert.Nil(t, pluginInfoTest.Result.Error)
	assert.Equal(t, filepath.Join(testOrchDir, "awsrunShellScript"), pluginInfoTest.Configuration.OrchestrationDirectory)
	assert.Equal(t, testS3Bucket, pluginInfoTest.Configuration.OutputS3BucketName)
	assert.Equal(t, filepath.Join(testS3Prefix, "awsrunShellScript"), pluginInfoTest.Configuration.OutputS3KeyPrefix)
	assert.Equal(t, testMessageID, pluginInfoTest.Configuration.MessageId)
	assert.Equal(t, testDocumentID, pluginInfoTest.Configuration.BookKeepingFileName)
	assert.Equal(t, testWorkingDir, pluginInfoTest.Configuration.DefaultWorkingDirectory)
}

func TestParseDocument_ValidMainSteps(t *testing.T) {
	mockLog := log.NewMockLog()

	testParserInfo := DocumentParserInfo{
		OrchestrationDir:  testOrchDir,
		S3Bucket:          testS3Bucket,
		S3Prefix:          testS3Prefix,
		MessageId:         testMessageID,
		DocumentId:        testDocumentID,
		DefaultWorkingDir: testWorkingDir,
	}

	var testDocContent contracts.DocumentContent
	validdocumentmainsteps := loadFile(t, "../framework/service/runcommand/mds/testdata/validcommand20.json")
	err := json.Unmarshal(validdocumentmainsteps, &testDocContent)
	if err != nil {
		assert.Error(t, err, "Error occured when trying to unmarshal validDocument")
	}
	pluginsInfo, err := ParseDocument(mockLog, &testDocContent, testParserInfo, nil)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(pluginsInfo))

	pluginInfoTest := pluginsInfo[0]
	assert.Nil(t, pluginInfoTest.Result.Error)
	assert.Equal(t, filepath.Join(testOrchDir, "test"), pluginInfoTest.Configuration.OrchestrationDirectory)
	assert.Equal(t, testS3Bucket, pluginInfoTest.Configuration.OutputS3BucketName)
	assert.Equal(t, filepath.Join(testS3Prefix, "awsrunShellScript"), pluginInfoTest.Configuration.OutputS3KeyPrefix)
	assert.Equal(t, testMessageID, pluginInfoTest.Configuration.MessageId)
	assert.Equal(t, testDocumentID, pluginInfoTest.Configuration.BookKeepingFileName)
	assert.Equal(t, testWorkingDir, pluginInfoTest.Configuration.DefaultWorkingDirectory)
}

func TestInitializeDocState_Valid(t *testing.T) {
	mockLog := log.NewMockLog()

	testParserInfo := DocumentParserInfo{
		OrchestrationDir:  testOrchDir,
		S3Bucket:          testS3Bucket,
		S3Prefix:          testS3Prefix,
		MessageId:         testMessageID,
		DocumentId:        testDocumentID,
		DefaultWorkingDir: testWorkingDir,
	}

	var testDocContent contracts.DocumentContent
	validdocumentruntimeconfig := loadFile(t, "../framework/service/runcommand/mds/testdata/validcommand12.json")
	err := json.Unmarshal(validdocumentruntimeconfig, &testDocContent)
	if err != nil {
		assert.Error(t, err, "Error occured when trying to unmarshal validDocument")
	}

	docState, err := InitializeDocState(mockLog, model.SendCommand, &testDocContent, model.DocumentInfo{}, testParserInfo, nil)

	assert.Nil(t, err)

	pluginInfo := docState.InstancePluginsInformation
	assert.Equal(t, model.SendCommand, docState.DocumentType)
	assert.Equal(t, "1.2", docState.SchemaVersion)
	assert.Equal(t, 1, len(pluginInfo))
	assert.Equal(t, filepath.Join(testOrchDir, "awsrunShellScript"), pluginInfo[0].Configuration.OrchestrationDirectory)
	assert.Equal(t, testS3Bucket, pluginInfo[0].Configuration.OutputS3BucketName)
	assert.Equal(t, filepath.Join(testS3Prefix, "awsrunShellScript"), pluginInfo[0].Configuration.OutputS3KeyPrefix)
	assert.Equal(t, testMessageID, pluginInfo[0].Configuration.MessageId)
	assert.Equal(t, testDocumentID, pluginInfo[0].Configuration.BookKeepingFileName)
	assert.Equal(t, testWorkingDir, pluginInfo[0].Configuration.DefaultWorkingDirectory)
}

func TestParseDocument_EmptyDocContent(t *testing.T) {
	mockLog := log.NewMockLog()
	testParserInfo := DocumentParserInfo{
		OrchestrationDir:  testOrchDir,
		S3Bucket:          testS3Bucket,
		S3Prefix:          testS3Prefix,
		MessageId:         testMessageID,
		DocumentId:        testDocumentID,
		DefaultWorkingDir: testWorkingDir,
	}

	var testDocContent contracts.DocumentContent
	testDocContent.SchemaVersion = "1.2"
	_, err := ParseDocument(mockLog, &testDocContent, testParserInfo, nil)

	assert.NotNil(t, err)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Unsupported schema format")
}

func TestParseDocument_Invalid(t *testing.T) {
	mockLog := log.NewMockLog()
	testParserInfo := DocumentParserInfo{
		OrchestrationDir:  testOrchDir,
		S3Bucket:          testS3Bucket,
		S3Prefix:          testS3Prefix,
		MessageId:         testMessageID,
		DocumentId:        testDocumentID,
		DefaultWorkingDir: testWorkingDir,
	}

	var testDocContent contracts.DocumentContent
	err := json.Unmarshal([]byte(invaliddocument), &testDocContent)
	assert.Nil(t, err)
	assert.NoError(t, err, "Error occured when trying to unmarshal invalidDocument")
	_, err = ParseDocument(mockLog, &testDocContent, testParserInfo, nil)

	assert.NotNil(t, err)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Unsupported schema format")
}

func TestParseDocument_InvalidSchema(t *testing.T) {
	mockLog := log.NewMockLog()
	testParserInfo := DocumentParserInfo{
		OrchestrationDir:  testOrchDir,
		S3Bucket:          testS3Bucket,
		S3Prefix:          testS3Prefix,
		MessageId:         testMessageID,
		DocumentId:        testDocumentID,
		DefaultWorkingDir: testWorkingDir,
	}
	var testDocContent contracts.DocumentContent
	invalidschema := loadFile(t, "testdata/schemaVersion9999.json")

	err := json.Unmarshal([]byte(invalidschema), &testDocContent)
	assert.Nil(t, err)
	assert.NoError(t, err, "Error occured when trying to unmarshal invalidschema")

	_, err = ParseDocument(mockLog, &testDocContent, testParserInfo, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Document with schema version 9999.0 is not supported by this version of ssm agent")
}

func TestParseDocument_ValidParameters(t *testing.T) {
	mockLog := log.NewMockLog()

	testParserInfo := DocumentParserInfo{
		OrchestrationDir:  testOrchDir,
		S3Bucket:          testS3Bucket,
		S3Prefix:          testS3Prefix,
		MessageId:         testMessageID,
		DocumentId:        testDocumentID,
		DefaultWorkingDir: testWorkingDir,
	}

	var testDocContent contracts.DocumentContent
	err := json.Unmarshal([]byte(parameterdocument), &testDocContent)
	assert.Nil(t, err)
	assert.NoError(t, err, "Error occured when trying to unmarshal validDocument")

	var testParams map[string]interface{}
	err = json.Unmarshal([]byte(testparameters), &testParams)
	assert.Nil(t, err)
	assert.NoError(t, err, "Error occured when trying to unmarshal testparameters")
	originalMessage, _ := jsonutil.Marshal(testDocContent)

	pluginsInfo, err := ParseDocument(mockLog, &testDocContent, testParserInfo, testParams)
	parsedMessage, _ := jsonutil.Marshal(testDocContent)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(pluginsInfo))

	pluginInfoTest := pluginsInfo[0]
	assert.Nil(t, pluginInfoTest.Result.Error)
	assert.Equal(t, filepath.Join(testOrchDir, "awsrunPowerShellScript"), pluginInfoTest.Configuration.OrchestrationDirectory)
	assert.Equal(t, testS3Bucket, pluginInfoTest.Configuration.OutputS3BucketName)
	assert.Equal(t, filepath.Join(testS3Prefix, "awsrunPowerShellScript"), pluginInfoTest.Configuration.OutputS3KeyPrefix)
	assert.Equal(t, testMessageID, pluginInfoTest.Configuration.MessageId)
	assert.Equal(t, testDocumentID, pluginInfoTest.Configuration.BookKeepingFileName)
	assert.Equal(t, testWorkingDir, pluginInfoTest.Configuration.DefaultWorkingDirectory)
	assert.NotEqual(t, parsedMessage, originalMessage)
}

func TestIsCrossPlatformEnabledForSchema20(t *testing.T) {
	var schemaVersion = "2.0"
	isCrossPlatformEnabled := isPreconditionEnabled(schemaVersion)

	// isCrossPlatformEnabled should be false for 2.0 document
	assert.False(t, isCrossPlatformEnabled)
}

func TestIsCrossPlatformEnabledForSchema22(t *testing.T) {
	var schemaVersion = "2.2"
	isCrossPlatformEnabled := isPreconditionEnabled(schemaVersion)

	// isCrossPlatformEnabled should be true for 2.2 document
	assert.True(t, isCrossPlatformEnabled)
}

func TestParseMessageWithParams(t *testing.T) {
	type testCase struct {
		Input       string
		OutputDoc   contracts.DocumentContent
		OutputParam map[string]interface{}
	}
	mockLog := log.NewMockLog()

	testParserInfo := DocumentParserInfo{
		OrchestrationDir:  testOrchDir,
		S3Bucket:          testS3Bucket,
		S3Prefix:          testS3Prefix,
		MessageId:         testMessageID,
		DocumentId:        testDocumentID,
		DefaultWorkingDir: testWorkingDir,
	}

	// generate test cases
	var testCases []testCase
	for _, msgFileName := range sampleMessageFiles {
		outputDoc, outputParam := loadMessageFromFile(t, msgFileName)
		fmt.Print(msgFileName)
		testCases = append(testCases, testCase{
			Input:       string(loadFile(t, msgFileName)),
			OutputDoc:   outputDoc,
			OutputParam: outputParam,
		})
	}

	// run tests
	for _, tst := range testCases {
		// call method
		origMessage, _ := jsonutil.Marshal(tst.OutputDoc)
		pluginsInfo, err := ParseDocument(mockLog, &tst.OutputDoc, testParserInfo, tst.OutputParam)
		parsedMessage, _ := jsonutil.Marshal(tst.OutputDoc)

		// check results
		fmt.Print(tst.OutputDoc)
		assert.Nil(t, err)
		assert.Equal(t, testS3Bucket, pluginsInfo[0].Configuration.OutputS3BucketName)
		assert.Equal(t, testMessageID, pluginsInfo[0].Configuration.MessageId)
		assert.Equal(t, testDocumentID, pluginsInfo[0].Configuration.BookKeepingFileName)
		assert.Equal(t, testWorkingDir, pluginsInfo[0].Configuration.DefaultWorkingDirectory)
		assert.NotEqual(t, origMessage, parsedMessage)
	}
}

func loadFile(t *testing.T, fileName string) (result []byte) {
	result, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	return
}

func loadMessageFromFile(t *testing.T, fileName string) (testDocContent contracts.DocumentContent, params map[string]interface{}) {
	b := loadFile(t, fileName)
	err := json.Unmarshal(b, &testDocContent)
	if err != nil {
		t.Fatal(err)
	}
	p := loadFile(t, "testdata/sampleMessageParameters.json")
	err = json.Unmarshal(p, &params)
	if err != nil {
		t.Fatal(err)
	}
	return testDocContent, params
}
