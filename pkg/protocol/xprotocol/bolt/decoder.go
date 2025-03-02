/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package bolt

import (
	"context"
	"encoding/binary"

	"mosn.io/api"
	"mosn.io/mosn/pkg/types"
	"mosn.io/pkg/variable"
	"mosn.io/pkg/buffer"
	"mosn.io/pkg/header"
)

func decodeRequest(ctx context.Context, data api.IoBuffer, oneway bool) (cmd interface{}, err error) {
	bytesLen := data.Len()
	bytes := data.Bytes()

	// 1. least bytes to decode header is RequestHeaderLen(22)
	if bytesLen < RequestHeaderLen {
		return
	}

	// 2. least bytes to decode whole frame
	classLen := binary.BigEndian.Uint16(bytes[14:16])
	headerLen := binary.BigEndian.Uint16(bytes[16:18])
	contentLen := binary.BigEndian.Uint32(bytes[18:22])

	frameLen := RequestHeaderLen + int(classLen) + int(headerLen) + int(contentLen)
	if bytesLen < frameLen {
		return
	}
	data.Drain(frameLen)

	// 3. decode header
	buf := bufferByContext(ctx)
	request := &buf.request

	cmdType := CmdTypeRequest
	if oneway {
		cmdType = CmdTypeRequestOneway
	}

	request.RequestHeader = RequestHeader{
		Protocol:   ProtocolCode,
		CmdType:    cmdType,
		CmdCode:    binary.BigEndian.Uint16(bytes[2:4]),
		Version:    bytes[4],
		RequestId:  binary.BigEndian.Uint32(bytes[5:9]),
		Codec:      bytes[9],
		Timeout:    int32(binary.BigEndian.Uint32(bytes[10:14])),
		ClassLen:   classLen,
		HeaderLen:  headerLen,
		ContentLen: contentLen,
	}
	request.Data = buffer.GetIoBuffer(frameLen)

	//4. copy data for io multiplexing
	request.Data.Write(bytes[:frameLen])
	request.rawData = request.Data.Bytes()

	// notice: read-only!!! do not modify the raw data!!!
	variable.Set(ctx, types.VarRequestRawData, request.rawData)

	//5. process wrappers: Class, Header, Content, Data
	headerIndex := RequestHeaderLen + int(classLen)
	contentIndex := headerIndex + int(headerLen)

	request.rawMeta = request.rawData[:RequestHeaderLen]
	if classLen > 0 {
		request.rawClass = request.rawData[RequestHeaderLen:headerIndex]
		request.Class = string(request.rawClass)
	}
	if headerLen > 0 {
		request.rawHeader = request.rawData[headerIndex:contentIndex]
		err = header.DecodeHeader(request.rawHeader, &request.BytesHeader)
	}
	if contentLen > 0 {
		request.rawContent = request.rawData[contentIndex:]
		request.Content = buffer.NewIoBufferBytes(request.rawContent)
	}
	return request, err
}

func decodeResponse(ctx context.Context, data api.IoBuffer) (cmd interface{}, err error) {
	bytesLen := data.Len()
	bytes := data.Bytes()

	// 1. least bytes to decode header is ResponseHeaderLen(20)
	if bytesLen < ResponseHeaderLen {
		return
	}

	// 2. least bytes to decode whole frame
	classLen := binary.BigEndian.Uint16(bytes[12:14])
	headerLen := binary.BigEndian.Uint16(bytes[14:16])
	contentLen := binary.BigEndian.Uint32(bytes[16:20])

	frameLen := ResponseHeaderLen + int(classLen) + int(headerLen) + int(contentLen)
	if bytesLen < frameLen {
		return
	}
	data.Drain(frameLen)

	// 3. decode header
	buf := bufferByContext(ctx)
	response := &buf.response

	response.ResponseHeader = ResponseHeader{
		Protocol:       ProtocolCode,
		CmdType:        CmdTypeResponse,
		CmdCode:        binary.BigEndian.Uint16(bytes[2:4]),
		Version:        bytes[4],
		RequestId:      binary.BigEndian.Uint32(bytes[5:9]),
		Codec:          bytes[9],
		ResponseStatus: binary.BigEndian.Uint16(bytes[10:12]),
		ClassLen:       classLen,
		HeaderLen:      headerLen,
		ContentLen:     contentLen,
	}
	response.Data = buffer.GetIoBuffer(frameLen)

	//TODO: test recycle by model, so we can recycle request/response models, headers also
	//4. copy data for io multiplexing
	response.Data.Write(bytes[:frameLen])
	response.rawData = response.Data.Bytes()

	// notice: read-only!!! do not modify the raw data!!!
	variable.Set(ctx, types.VarResponseRawData, response.rawData)

	//5. process wrappers: Class, Header, Content, Data
	headerIndex := ResponseHeaderLen + int(classLen)
	contentIndex := headerIndex + int(headerLen)

	response.rawMeta = response.rawData[:ResponseHeaderLen]
	if classLen > 0 {
		response.rawClass = response.rawData[ResponseHeaderLen:headerIndex]
		response.Class = string(response.rawClass)
	}
	if headerLen > 0 {
		response.rawHeader = response.rawData[headerIndex:contentIndex]
		err = header.DecodeHeader(response.rawHeader, &response.BytesHeader)
	}
	if contentLen > 0 {
		response.rawContent = response.rawData[contentIndex:]
		response.Content = buffer.NewIoBufferBytes(response.rawContent)
	}
	return response, err
}
